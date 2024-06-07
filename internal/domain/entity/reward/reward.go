package reward

import (
	"time"

	"github.com/google/uuid"

	"github.com/r1der/epos/internal/domain/entity/position"
	"github.com/r1der/epos/internal/domain/values"
)

type Reward struct {
	id        uuid.UUID
	pos       *position.Position
	amount    values.Amount
	createdAt time.Time
}

func (r *Reward) ID() uuid.UUID                { return r.id }
func (r *Reward) Position() *position.Position { return r.pos }
func (r *Reward) Amount() values.Amount        { return r.amount }
func (r *Reward) CreatedAt() time.Time         { return r.createdAt }

func New(pos *position.Position, amount values.Amount) *Reward {
	return &Reward{
		id:        uuid.New(),
		pos:       pos,
		amount:    amount,
		createdAt: time.Now(),
	}
}
