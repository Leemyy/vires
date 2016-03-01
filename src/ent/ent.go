// Package ent contains game entities
// present on a game field.
package ent

import (
	"math"
	"time"

	"github.com/mhuisi/vires/src/vec"
)

type (
	// Vires represents a game unit
	Vires int
	// ID represents any kind of unique identifier value
	ID int
)

// Player represents a player
// playing the game.
type Player struct {
	id    ID
	cells int
}

// NewPlayer creates a new player
// with the specified ID and
// an amount of cells of 1.
func NewPlayer(id ID) Player {
	return Player{id, 1}
}

// ID gets the id of the player.
func (p Player) ID() ID {
	return p.id
}

// Cells gets the amount of cells of the player.
func (p Player) Cells() int {
	return p.cells
}

// IsDead returns whether the player has no cells left.
func (p Player) IsDead() bool {
	return p.cells <= 0
}

// Circle represents a 2D circle.
type Circle struct {
	Location vec.V
	Radius   float64
}

// Cell represents a cell on the field.
type Cell struct {
	id       ID
	force    float64
	capacity Vires
	// [Replication] = vires/cycle
	replication Vires
	stationed   Vires
	// may be nil
	owner *Player
	body  Circle
}

// NewCell creates a new cell with the specified id,
// the specified force, which determines its capacity, its
// replication and its radius and loc, which is the
// location of the cell.
func NewCell(id ID, force float64, loc vec.V) *Cell {
	return &Cell{
		id:          id,
		force:       force,
		capacity:    capacity(force),
		replication: neutralReplication(force),
		stationed:   0,
		body:        Circle{loc, cellRadius(force)},
	}
}

// ID gets the id of the cell.
func (c *Cell) ID() ID {
	return c.id
}

// Capacity gets the limit for stationed vires
// of this cell.
func (c *Cell) Capacity() Vires {
	return c.capacity
}

// Replication gets the rate at which
// vires in this cell replicate, ie
// how many vires are produced each cycle.
func (c *Cell) Replication() Vires {
	return c.replication
}

// Stationed gets the amount of vires
// stationed in this cell.
func (c *Cell) Stationed() Vires {
	return c.stationed
}

// Owner gets the owner of this cell.
// May be nil if the cell is neutral.
func (c *Cell) Owner() *Player {
	if c.owner == nil {
		return nil
	}
	clone := *c.owner
	return &clone
}

func (c *Cell) OwnerID() ID {
	o := c.Owner()
	if o == nil {
		return 0
	}
	return o.ID()
}

// Body gets the radius and the location
// of the cell.
func (c *Cell) Body() Circle {
	return c.body
}

func capacity(force float64) Vires {
	// placeholder, needs testing
	return Vires(math.Pi * sq(force) / 200)
}

func replication(force float64) Vires {
	// placeholder, needs testing
	return Vires(force / 5)
}

func neutralReplication(force float64) Vires {
	return replication(force) / 2
}

func cellRadius(force float64) float64 {
	// placeholder, needs testing
	return force
}

// Merge adds the specified amount
// of vires into this cell
// and limits the merge if
// the new amount of stationed vires
// is smaller than 0 or larger than the
// capacity.
func (c *Cell) Merge(n Vires) {
	newStationed := c.stationed + n
	switch {
	case newStationed < 0:
		c.stationed = 0
	case newStationed > c.capacity:
		c.stationed = c.capacity
	default:
		c.stationed = newStationed
	}
}

// Replicate runs a single replication
// cycle.
func (c *Cell) Replicate() {
	c.Merge(c.replication)
}

// IsDead returns whether the cell is dead.
func (c *Cell) IsDead() bool {
	return c.stationed <= 0
}

// IsNeutral returns whether the cell has an owner.
func (c *Cell) IsNeutral() bool {
	return c.owner == nil
}

// SetOwner sets the owner of this cell,
// removes the cell from the original owner
// and adds the cell to the new owner.
func (c *Cell) SetOwner(o Player) {
	if c.IsNeutral() {
		c.replication = replication(c.force)
	} else {
		c.owner.cells--
	}
	o.cells++
	c.owner = &o
}

// Neutralize resets the owner of this cell
// and removes the cell from its original owner.
func (c *Cell) Neutralize() {
	if c.IsNeutral() {
		return
	}
	c.owner.cells--
	c.owner = nil
	c.replication = neutralReplication(c.force)
}

func radius(n Vires) float64 {
	return 10 * math.Sqrt(float64(n)/math.Pi)
}

func speed(radius float64) float64 {
	if radius == 0 {
		return 0
	}
	return 3000 / radius
}

// Move creates a movement which describes
// a movement from this cell to the specified
// target cell and uses the specified id.
func (src *Cell) Move(mvid ID, tgt *Cell) *Movement {
	moving := src.stationed / 2
	start := src.body.Location
	r := radius(moving)
	mov := &Movement{
		id:         mvid,
		owner:      *src.owner,
		moving:     moving,
		target:     tgt,
		body:       Circle{start, r},
		lastTime:   time.Now(),
		direction:  vec.Scale(vec.SubV(tgt.body.Location, start), speed(r)),
		collisions: map[*Movement]func(){},
	}
	src.Merge(-moving)
	return mov
}

// Movement represents vires that are
// moving in between cells.
type Movement struct {
	id       ID
	owner    Player
	moving   Vires
	target   *Cell
	body     Circle
	lastTime time.Time
	// |Direction| = v, [v] = points/s
	direction  vec.V
	collisions map[*Movement]func()
	Stop       func()
}

// ID gets the id of the movement.
func (m *Movement) ID() ID {
	return m.id
}

// Owner gets the owner of the movement,
// ie the player that sent the movement.
func (m *Movement) Owner() Player {
	return m.owner
}

// Moving gets the amount of vires present
// in this movement.
func (m *Movement) Moving() Vires {
	return m.moving
}

// Target gets the cell towards which this
// movement moves.
// The return value should not be mutated.
func (m *Movement) Target() *Cell {
	return m.target
}

// Body gets the radius and the current location
// of this movement.
func (m *Movement) Body() Circle {
	return m.body
}

// Direction gets the direction the
// movement is moving in.
// The abs() of the returned vector
// is the speed of the movement in
// points/s.
func (m *Movement) Direction() vec.V {
	return m.direction
}

// Collisions gets the movements this movement collides with.
// The value of the returned map is a function to
// stop the respective collision from happening.
// The return value should not be mutated.
func (m *Movement) Collisions() map[*Movement]func() {
	return m.collisions
}

// AddCollision adds a movement this movement collides with
// to the movement. stopCollision is called when
// the movement is stopped.
func (m *Movement) AddCollision(m2 *Movement, stopCollision func()) {
	m.collisions[m2] = stopCollision
}

// ClearCollisions stops all collisions associated with this movement
// and removes this movement from all the movements this movement
// is colliding with.
func (m *Movement) ClearCollisions() {
	for m2, stop := range m.collisions {
		stop()
		delete(m.collisions, m2)
		delete(m2.collisions, m)
	}
}

// Merge adds the amount of specified vires
// into this cell and rescales
// its speed and its radius dependent
// on the new amount of vires.
func (m *Movement) Merge(n Vires) {
	newMoving := m.moving + n
	if newMoving < 0 {
		newMoving = 0
	}
	m.moving = newMoving
	r := radius(newMoving)
	m.direction = vec.Scale(m.direction, speed(r))
	m.body.Radius = r
}

// Kill sets the amount of vires of this movement to 0.
func (m *Movement) Kill() {
	m.Merge(-m.moving)
}

// Conflict executes a collision
// with the target cell, merging
// the vires into the target cell
// if it is a friendly cell or
// or removing the vires from the
// target cell if it is an enemy cell.
// Cell owners are transferred
// if the movement is strong enough
// to take over the cell.
func (m *Movement) Conflict() {
	// after conflict, set moving to 0
	defer m.Kill()
	tgt := m.target
	attacker := m.owner
	defid := tgt.OwnerID()
	// same player, friendly units are merged into the cell
	if attacker.ID() == defid {
		tgt.Merge(m.moving)
	} else {
		tgt.Merge(-m.moving)
		// cell died, change owner
		if tgt.IsDead() {
			tgt.SetOwner(attacker)
		}
	}
}

// Collide executes a collision with a
// movement, merging the movements
// if they have the same target and
// the same owner or substracting vires
// from each other when the owners are
// different.
// The set of collisions in a movement
// is not modified by this method.
func (m *Movement) Collide(m2 *Movement) {
	m.UpdatePosition()
	m2.UpdatePosition()
	// merge movements if two movements with the same owner and the same target collide
	if m.owner.ID() == m2.owner.ID() {
		mt := m.target
		m2t := m2.target
		if mt == m2t {
			// collision with friendly movement
			d1 := vec.Dist(mt.Body().Location, m.Body().Location)
			d2 := vec.Dist(m2t.Body().Location, m2.Body().Location)
			// merge movement that is further away into
			// movement that is closer
			if d2 < d1 {
				m, m2 = m2, m
			}
			m.Merge(m2.moving)
			m2.Kill()
			return
		}
		// no collision, friendly movements cross each other
		return
	}
	// standard collision
	m.Merge(-m2.moving)
	m2.Merge(-m.moving)
}

// IsDead returns whether the movement is dead.
func (m *Movement) IsDead() bool {
	return m.moving <= 0
}

func sq(v float64) float64 {
	return v * v
}

func collideIn(m1 *Movement, m2 *Movement) (float64, bool) {
	// concept:
	// we treat one movement relative to the other movement
	// and then calculate the times at which the path of the smaller
	// movement intersects the circle bounds of the larger movement.
	// because we treat both movements relative to each other,
	// the center of the larger movement is at (0, 0).
	b1 := m1.body
	b2 := m2.body
	p := vec.SubV(b1.Location, b2.Location)
	v := vec.SubV(m1.direction, m2.direction)
	r := math.Max(b1.Radius, b2.Radius)
	d := vec.Unit(v)
	tempP := vec.Dot(p, d)
	tempR := sq(tempP) - sq(p.X) - sq(p.Y) + sq(r)
	if tempR < 0 {
		// no collision
		return math.NaN(), false
	}
	tr := math.Sqrt(tempR)
	t1 := tr - tempP
	t2 := -tr - tempP
	switch {
	case t1 > 0 && t2 > 0:
		// collision is at t
		t := math.Min(t1, t2) / vec.Abs(v)
		return t, true
	case t1 <= 0 && t2 <= 0:
		// no collision
		return math.NaN(), false
	default:
		// collision is now
		return 0, true
	}
}

func at(in float64) time.Time {
	return time.Now().Add(time.Duration(in * float64(time.Second)))
}

func (m *Movement) UpdatePosition() {
	now := time.Now()
	m.body.Location = vec.AddV(m.body.Location, vec.Mul(m.direction, float64(now.Sub(m.lastTime))/float64(time.Second)))
	m.lastTime = now
}

// CollidesWith checks if a collision with the
// specified movement occurs and returns
// at what time it occurs.
func (m *Movement) CollidesWith(m2 *Movement) (collideAt time.Time, collides bool) {
	if m.ID() == m2.ID() {
		return time.Now(), false
	}
	// movements where the owner is the same
	// but the target isn't don't collide;
	// the movements just pass each other
	if !(m.owner == m2.owner && m.target != m2.target) {
		m.UpdatePosition()
		m2.UpdatePosition()
		in, collides := collideIn(m, m2)
		return at(in), collides
	}
	return time.Now(), false
}

// ConflictAt returns at what time a
// collision with a cell occurs.
func (m *Movement) ConflictAt() time.Time {
	defender := m.target
	speed := vec.Abs(m.direction)
	dist := vec.Abs(vec.SubV(defender.body.Location, m.body.Location))
	delay := dist / speed
	return at(delay)
}
