package usecase

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/sirupsen/logrus"

	"github.com/r1der/epos/internal/domain/entity/order"
	"github.com/r1der/epos/internal/domain/entity/pool"
	"github.com/r1der/epos/internal/domain/entity/position"
	"github.com/r1der/epos/internal/domain/entity/project"
	"github.com/r1der/epos/internal/domain/entity/reward"
	"github.com/r1der/epos/internal/domain/ports"
	"github.com/r1der/epos/internal/domain/values"
)

type ProjectExecutor interface {
	Execute(context.Context, *project.Project) error
}

type projectExecutor struct {
	poolManager     pool.Manager
	positionManager position.Manager
	projectManager  project.Manager
	orderManager    order.Manager
	rewardManger    reward.Manager
	balance         ports.Balance
}

func NewProjectExecutor(
	poolManager pool.Manager,
	positionManager position.Manager,
	projectManager project.Manager,
	orderManager order.Manager,
	rewardManger reward.Manager,
	balance ports.Balance,
) ProjectExecutor {
	return &projectExecutor{
		poolManager:     poolManager,
		positionManager: positionManager,
		projectManager:  projectManager,
		orderManager:    orderManager,
		rewardManger:    rewardManger,
		balance:         balance,
	}
}

// Execute runs a strategy
func (svc *projectExecutor) Execute(ctx context.Context, proj *project.Project) error {
	can, err := svc.canBeExecuted(ctx, proj)
	if err != nil {
		return fmt.Errorf("can the project be executed: %w", err)
	}
	if !can {
		return nil
	}

	// смотрим, есть ли открытые позиции
	openPositions, err := svc.positionManager.GetOpenPositions(ctx, proj)
	if err != nil {
		return fmt.Errorf("position manager: open positions: %w", err)
	}
	logrus.Printf("open positions: %d", len(openPositions))

	// если нет активных позиций, запускаем новую
	if len(openPositions) > 0 {
		return svc.open(ctx, proj)
	}

	// @todo сделать возможность держать несколько позиций
	// проверяем, позиция в рендже или нет
	openPosition := openPositions[0]
	return svc.check(ctx, proj, openPosition)
}

// open creates a new position
func (svc *projectExecutor) open(ctx context.Context, proj *project.Project) error {
	// определяем диапазон для позиции
	// диапазон зависит от точки входа и указанной волатильности
	pricesRange, err := svc.poolManager.CalculatePositionRange(ctx, proj.Pool(), proj.RangeVolatility())
	if err != nil {
		return fmt.Errorf("calculate range: %w", err)
	}
	log.Printf("position price range calculated: lower: %f, current: %f, upper: %f",
		pricesRange.LowerPrice, pricesRange.InitialPrice, pricesRange.UpperPrice)

	// определяем сумму инвестиций
	// сумма делиться на кол-во разрешенных открытых позиций
	investment := proj.Investments().Div(proj.ActivePositions())
	logrus.Printf("selected investment: %d %s (%f %s)",
		investment.Value(), investment.Token(), investment.HumanValue(), investment.Token())

	// рассчитываем примерный размер позиции в зависимости от суммы инвестиций
	var baseAmount, quoteAmount values.Amount

	pair := proj.Pool().Pair()
	half := investment.Div(2)

	if investment.Token().Eq(pair.BaseToken()) { // investment in base asset
		baseAmount = values.NewAmount(pair.BaseToken(), half.Value())
		quoteAmount = values.NewAmount(pair.QuoteToken(), half.Mul(pricesRange.InitialPrice).Value())
	} else if investment.Token().Eq(pair.QuoteToken()) { // investment in quote asset
		baseAmount = values.NewAmount(pair.BaseToken(), half.Div(pricesRange.InitialPrice).Value())
		quoteAmount = values.NewAmount(pair.QuoteToken(), half.Value())
	} else {
		// @todo реализовать в будущем ZapIn токенов отличных от активов пула
		return fmt.Errorf("invalid investment token")
	}
	logrus.Printf("est. position amounts calculated: base amount: %d %s (%f %s), quote amount: %d %s (%f %s)",
		baseAmount.Value(), baseAmount.Token(), baseAmount.HumanValue(), baseAmount.Token(),
		quoteAmount.Value(), quoteAmount.Token(), quoteAmount.HumanValue(), quoteAmount.Token())

	// высчитываем точный размер позиции с учетом текущей цены в пуле
	amounts, err := svc.poolManager.CalculatePositionAmounts(ctx, pricesRange, baseAmount, quoteAmount)
	if err != nil {
		return fmt.Errorf("calculate position amounts: %w", err)
	}

	baseAmount, quoteAmount = amounts.BaseAmount, amounts.QuoteAmount
	logrus.Printf("position amounts adjusted: base amount: %d %s (%f %s), quote amount: %d %s (%f %s)",
		baseAmount.Value(), baseAmount.Token(), baseAmount.HumanValue(), baseAmount.Token(),
		quoteAmount.Value(), quoteAmount.Token(), quoteAmount.HumanValue(), quoteAmount.Token())

	// получаем средства на балансе выбранных активов
	baseBalance, err := svc.balance.Get(ctx, proj.Wallet(), pair.BaseToken())
	if err != nil {
		return fmt.Errorf("get base token balance: %w", err)
	}
	quoteBalance, err := svc.balance.Get(ctx, proj.Wallet(), pair.QuoteToken())
	if err != nil {
		return fmt.Errorf("get quote token balance: %w", err)
	}
	logrus.Printf("balances loaded: base %d %s (%f %s), quote: %d %s (%f %s)",
		baseBalance.Value(), baseBalance.Token(), baseBalance.HumanValue(), baseBalance.Token(),
		quoteBalance.Value(), quoteBalance.Token(), quoteBalance.HumanValue(), quoteBalance.Token())

	// проверяем, можем ли мы открыть позицию с нашими активами и сделать своп при необходимости с учетом слиппаджа
	can, swapData := svc.canBeOpened(pricesRange.InitialPrice, baseAmount, quoteAmount, baseBalance, quoteBalance, proj.Slippage())
	if !can {
		if err = svc.projectManager.Deactivate(ctx, proj, project.NotEnoughFunds); err != nil {
			return fmt.Errorf("deactivate project: %w", err)
		}
		log.Printf("project deactivated: %s, reason: %s", proj.ID().String(), proj.InactiveReason())
		return nil
	}

	// делаем своп недостающих активов
	if swapData != nil {
		if err = svc.swap(ctx, proj, swapData); err != nil {
			return fmt.Errorf("swap: %w", err)
		}
	}

	// открываем новую позицию
	pos, err := svc.positionManager.Open(ctx, &position.OpenPositionInput{
		Project:     proj,
		InitPrice:   pricesRange.InitialPrice,
		LowerPrice:  pricesRange.LowerPrice,
		UpperPrice:  pricesRange.UpperPrice,
		BaseAmount:  baseAmount,
		QuoteAmount: quoteAmount,
	})
	if err != nil {
		return fmt.Errorf("open position: %w", err)
	}

	logrus.Printf("position %s on %s/%s openned: %s",
		pos.Pool().Pair(), pos.Pool().Network(), pos.Pool().Protocol(), pos.Address())

	return nil
}

// check checks the current position range
func (svc *projectExecutor) check(ctx context.Context, proj *project.Project, pos *position.Position) error {
	logrus.Debugf("start of checking the current position")

	// получаем актуальную информацию о пуле
	p, err := svc.poolManager.Get(ctx, proj.Pool().Network(), proj.Pool().Protocol(), proj.Pool().Pair(), proj.Pool().Fee())
	if err != nil {
		return err
	}
	logrus.Printf("pool %s updated: last price: %f", p.Pair(), p.LastPrice())

	if err = svc.positionManager.Actualize(ctx, pos); err != nil {
		return fmt.Errorf("actualize position: %w", err)
	}
	logrus.Printf(
		"position %s [%f] on %s/%s actualized: base amount: %d %s (%f %s), quote amount: %d %s (%f %s), base accrued fees: %d %s (%f %s), quote accrued fees: %d %s (%f %s): ",
		pos.Pool().Pair(), pos.Pool().Fee(), pos.Pool().Network(), pos.Pool().Protocol(),
		pos.CurrentBaseAmount().Value(), pos.CurrentBaseAmount().Token(), pos.CurrentBaseAmount().HumanValue(), pos.CurrentBaseAmount().Token(),
		pos.CurrentQuoteAmount().Value(), pos.CurrentQuoteAmount().Token(), pos.CurrentQuoteAmount().HumanValue(), pos.CurrentQuoteAmount().Token(),
		pos.CurrentBaseAccruedFees().Value(), pos.CurrentBaseAccruedFees().Token(), pos.CurrentBaseAccruedFees().HumanValue(), pos.CurrentBaseAccruedFees().Token(),
		pos.CurrentQuoteAccruedFees().Value(), pos.CurrentQuoteAccruedFees().Token(), pos.CurrentQuoteAccruedFees().HumanValue(), pos.CurrentQuoteAccruedFees().Token(),
	)

	if pos.IsInRange() {
		logrus.Printf("position %s [%f] on %s/%s is in-range: price %f [%f - %f]",
			pos.Pool().Pair(), pos.Pool().Fee(), pos.Pool().Network(), pos.Pool().Protocol(),
			pos.CurrentPrice(), pos.LowerPrice(), pos.UpperPrice())
		return nil
	}

	// закрываем позицию
	if err = svc.close(ctx, pos); err != nil {
		return fmt.Errorf("close position: %w", err)
	}
	logrus.Printf("position %s [%f] on %s/%s closed",
		pos.Pool().Pair(), pos.Pool().Fee(), pos.Pool().Network(), pos.Pool().Protocol())

	return nil
}

// close закрывает позицию (уменьшение ликвидности позиции и сбор всех вознаграждений)
func (svc *projectExecutor) close(ctx context.Context, pos *position.Position) error {
	if err := svc.positionManager.Close(ctx, pos); err != nil {
		return fmt.Errorf("close position: %w", err)
	}

	logrus.Printf("position %s on %s/%s closed",
		pos.Pool().Pair(), pos.Pool().Network(), pos.Pool().Protocol())

	amounts, err := svc.positionManager.CollectRewards(ctx, pos)
	if err != nil {
		return fmt.Errorf("collect rewards: %w", err)
	}

	for _, r := range amounts {
		logrus.Printf("reward collected: %d %s (%f %s)",
			r.Value(), r.Token(), r.HumanValue(), r.Token())
	}

	if _, err = svc.rewardManger.Add(ctx, pos, amounts...); err != nil {
		return fmt.Errorf("add rewards: %w", err)
	}

	logrus.Printf("rewards for position %s on %s/%s collected",
		pos.Pool().Pair(), pos.Pool().Network(), pos.Pool().Protocol())

	return svc.recalculateProjectWorth(ctx, pos)
}

// exchange swaps assets
func (svc *projectExecutor) swap(ctx context.Context, proj *project.Project, data *SwapData) error {
	logrus.Debugf("start of swapping tokens")

	ord, err := svc.orderManager.New(ctx, &order.NewOrderInput{
		Project:   proj,
		AmountIn:  data.AmountIn,
		AmountOut: data.AmountOut,
		Price:     data.Price,
	})
	if err != nil {
		return fmt.Errorf("new order: %w", err)
	}
	log.Printf("order of swapping %d %s (%f %s) => %d %s (%f %s) by price %f created: %s",
		ord.AmountIn().Value(), ord.AmountIn().Token(), ord.AmountIn().HumanValue(), ord.AmountIn().Token(),
		ord.AmountOut().Value(), ord.AmountOut().Token(), ord.AmountOut().HumanValue(), ord.AmountOut().Token(),
		ord.FilledPrice(), ord.Address())

	return nil
}

// canBeExecuted checks the project requirements and stop-loss
func (svc *projectExecutor) canBeExecuted(ctx context.Context, proj *project.Project) (bool, error) {
	if proj.IsInactive() {
		return false, nil
	}

	// проверяем газ в сети
	gas, err := svc.balance.Get(ctx, proj.Wallet(), proj.Wallet().NativeToken())
	if err != nil {
		return false, fmt.Errorf("get native balance: %w", err)
	}
	if gas.IsZero() {
		if err = svc.projectManager.Deactivate(ctx, proj, project.NotEnoughGas); err != nil {
			return false, fmt.Errorf("deactivate project: %w", err)
		}
		log.Printf("project deactivated: %s, reason: %s", proj.ID().String(), proj.InactiveReason())
		return false, nil

	}

	stopLoss := values.NewAmount(proj.Investments().Token(), proj.Investments().Mul(proj.StopLoss().Float()))
	limit := proj.Investments().Sub(stopLoss)

	if proj.CurrentValue().Value().Cmp(limit.Value()) < 0 {
		if err = svc.projectManager.Deactivate(ctx, proj, project.StopLoss); err != nil {
			return false, fmt.Errorf("deactivate project: %w", err)
		}
		log.Printf("project deactivated: %s, reason: %s", proj.ID().String(), proj.InactiveReason())
		return false, nil
	}

	return true, nil
}

type SwapData struct {
	AmountIn  values.Amount
	AmountOut values.Amount
	Price     *big.Float
}

// canBeOpened checks whether there are enough current assets to open a position (taking into account swap if necessary)
func (svc *projectExecutor) canBeOpened(price *big.Float, baseAmount, quoteAmount, baseBalance, quoteBalance values.Amount, slippage values.Percent) (bool, *SwapData) {
	if baseBalance.Value().Cmp(baseAmount.Value()) == -1 && quoteBalance.Value().Cmp(quoteAmount.Value()) == -1 {
		// обоих активов на балансе меньше чем нужно для открытия позиции
		return false, nil
	} else if baseBalance.Value().Cmp(baseAmount.Value()) == 1 && quoteBalance.Value().Cmp(quoteAmount.Value()) == 1 {
		// обоих активов на балансе достаточно для открытия позиции
		return true, nil

	}

	var swapData *SwapData

	if baseBalance.Value().Cmp(baseAmount.Value()) == -1 { // меньше базового актива
		delta := baseAmount.Sub(baseBalance)
		deltaCost := values.NewAmount(quoteAmount.Token(), delta.Mul(price))
		deltaCostSlippage := deltaCost.Mul(slippage)
		deltaCostWithSlippage := deltaCost.Add(deltaCostSlippage)

		// корректирующего актива должно хватить на своп недостающего базового актива + слиппадж
		// и на саму позицию корректирующего актива
		if quoteBalance.Value().Cmp(quoteAmount.Add(deltaCostWithSlippage).Value()) == -1 {
			return false, nil
		}

		swapData = &SwapData{
			Price:     price,
			AmountIn:  deltaCostWithSlippage,
			AmountOut: delta,
		}
	} else { // меньше корректирующего актива
		delta := quoteAmount.Sub(quoteBalance)
		deltaCost := values.NewAmount(baseAmount.Token(), delta.Div(price))
		deltaCostSlippage := deltaCost.Mul(slippage)
		deltaCostWithSlippage := deltaCost.Add(deltaCostSlippage)

		// базового актива должно хватить на своп недостающего корректирующего актива + слиппадж
		// и на саму позицию базового актива
		if baseBalance.Value().Cmp(baseAmount.Add(deltaCostWithSlippage).Value()) == -1 {
			return false, nil
		}

		swapData = &SwapData{
			Price:     price,
			AmountIn:  deltaCostWithSlippage,
			AmountOut: delta,
		}
	}

	return true, swapData
}

func (svc *projectExecutor) recalculateProjectWorth(ctx context.Context, pos *position.Position) error {
	// стоимость активов которые вывели из пула по текущей цене
	// + стоимость заработанных комиссий

	proj := pos.Project()

	worth := values.NewAmount(proj.Investments().Token(), 0)

	currentPrice := pos.CurrentPrice()
	baseAmount := pos.OutputBaseAmount()
	quoteAmount := pos.OutputQuoteAmount()

	// получаем стоимость активов пула
	if worth.Token().Eq(baseAmount.Token()) {
		worth = worth.Add(baseAmount)
		worth = worth.Add(values.NewAmount(worth.Token(), quoteAmount.Div(currentPrice)))
	} else {
		worth = worth.Add(quoteAmount)
		worth = worth.Add(values.NewAmount(worth.Token(), baseAmount.Mul(currentPrice)))
	}
	logrus.Printf("position assets cost: %d %s (%f %s)",
		worth.Value(), worth.Token(), worth.HumanValue(), worth.Token())

	// получаем стоимость заработанных комиссий
	rewards, err := svc.rewardManger.GetPositionRewards(ctx, pos)
	if err != nil {
		return fmt.Errorf("get reward for position: %w", err)
	}

	// invest  => ETH
	// reward  => USDC ()

	rewardsWorth := values.NewAmount(worth.Token(), 0)
	for _, r := range rewards {
		if worth.Token().Eq(baseAmount.Token()) && r.Amount().Token().Eq(quoteAmount.Token()) {
			cost := values.NewAmount(worth.Token(), r.Amount().Div(currentPrice))
			rewardsWorth = rewardsWorth.Add(cost)
		} else if worth.Token().Eq(quoteAmount.Token()) && r.Amount().Token().Eq(baseAmount.Token()) {
			cost := values.NewAmount(worth.Token(), r.Amount().Mul(currentPrice))
			rewardsWorth = rewardsWorth.Add(cost)
		} else {
			rewardsWorth = rewardsWorth.Add(r.Amount())
		}
	}
	logrus.Printf("position rewards cost: %d %s (%f %s)",
		rewardsWorth.Value(), rewardsWorth.Token(), rewardsWorth.HumanValue(), rewardsWorth.Token())

	worth = worth.Add(rewardsWorth)
	logrus.Printf("total cost: %d %s (%f %s)",
		worth.Value(), worth.Token(), worth.HumanValue(), worth.Token())

	if err = svc.projectManager.UpdateWorth(ctx, proj, worth); err != nil {
		return fmt.Errorf("update project worth: %w", err)
	}

	return nil
}
