settings = 
	minZoom: 1
	maxZoom: 100
	zoomSpeed: 0.2
	factorLight: 0.3
	factorDark: 0.4
	indexNeutral: 10
	indexSelf: 30
	indexOther: 20
	indexMarker: 50
	offsetMovement: -100
	gauge: 0.1

vires =
	room: 0
	time: 0
	timePrev: 0
	timeDelta: 0
	Self: 0
	states:
		loading: {}
		lobby: {}
		match: {}
		noConnection: {}
		debug: {}
	#Currently active state
	active: null
	next:
		state: null
		data: null

	init: ->
		@active = @states.loading
		@active.load()
		return

	load: (stateName, data)->
		@next.state = @states[stateName]
		@next.data = data


vires.states.match = 

	players: null
	cells: null
	movements: null
	#Faster lookup of cells, needs to be sorted by x-coordinates
	lookup: null
	#Cell Markers
	selection: null
	targetMarker: null
	target: null
	#Synchronized random number generator
	random: null

	timeStart: 0
	fieldSize: vec2.fromValues(800, 800)
	spectating: false
	maxCellSize: 1
	minCellSize: 10000

	cameraDrag: false
	cameraStart: null


	load: (Field)->
		#Delete data of previous games
		server = new Player(0)
		server.color = gfx.makeColor(0)
		@players = { 0: server }
		@cells = new Array(Field.Cells.length)
		@lookup = new Array(Field.Cells.length)
		@movements = { }
		@selection = { }

		@timeStart = vires.time
		@targetMarker = new Primitive(vec2.create(), gfx.mesh.target, gfx.material.marker, settings.indexMarker)
		@targetMarker.unlink()

		#Set field size
		@fieldSize = vec2.fromValues(Field.Size.X, Field.Size.Y)

		#Initialize Field
		for cellData in Field.Cells
			@cells[cellData.ID] = new Cell(cellData)
			radius = cellData.Body.Radius
			if (radius > @maxCellSize)
				@maxCellSize = radius
			if (radius < @minCellSize)
				@minCellSize = radius

		#Prepare cell lookup
		@lookup = @cells.slice(0)
		@lookup.sort( (a, b)->
			return a.Pos[x]-b.Pos[x]
		)

		#Set up pseudo random number generator.
		#Will produce the same set of random numbers
		# for all players in a match.
		@random = new Random(vires.room)
		@random.next()
		@random.seed *= Math.floor(@cells[0].Radius)
		@random.next()
		@random.seed += Math.floor(@cells[0].Pos[0] ** @cells[0].Pos[1])
		@random.next()

		#Produce a shuffled set of colors to assign to the players
		palette = gfx.color.slice(1)
		@random.shuffle(palette)

		@spectating = true
		firstCell = null
		for i in [0...Field.StartCells.length]
			start = Field.StartCells[i]
			owner = new Player(start.Owner)
			owner.color = palette[i%palette.length]
			@players[start.Owner] = owner
			@cells[start.Cell].switchOwner(owner)

			if (owner.ID == vires.Self)
				@spectating = false
				firstCell = @cells[start.Cell]

		settings.minZoom = 800 / @fieldSize[1]
		settings.maxZoom = 200 / @minCellSize 
		if (@spectating)
			vec2.set(gfx.camera.pos, @fieldSize[x]/2, @fieldSize[y]/2)
			gfx.camera.zoom = gfx.width / @fieldSize[x]
		else
			vec2.set(gfx.camera.pos, firstCell.Pos[x], firstCell.Pos[y])
			gfx.camera.zoom = 100 / firstCell.Radius
		return

	unload: ->
		gfx.material.cell.clear()
		gfx.material.movement.clear()
		gfx.material.marker.clear()
		return

	digestInput: ->
		#if(input.leftPressed)
		#	mouse = input.cursor
		#	height = vires.time/10000
		#	console.log(height)
		#	new Primitive(mouse, gfx.mesh.round, gfx.color[Math.floor(Math.random()*gfx.color.length)], height)

		#Cell marking
		if (input.left && !@spectating)
			hover = @cellAt(input.cursor)
			if (hover?)
				#Cursor is over a Cell
				if (hover.Owner == vires.Self)
						#Hovered Cell is owned by the Player
						#Save it as selected
						@selection[hover.ID] = hover
				if !(@target?)
					#No Cell is currently marked as target
					#Mark hovered Cell as target
					@target = hover
					@targetMarker.pos = hover.Pos
					@targetMarker.scale = hover.Radius
					@targetMarker.link()
					hover.unmark()
				else if (@target.ID != hover.ID)
					#Another Cell is marked as target
					if (@target.Owner == vires.Self)
						#That Cell was owned by the Player
						#Place source marker
						@target.mark()
					#Mark hovered Cell as target
					@target = hover
					@targetMarker.pos = hover.Pos
					@targetMarker.scale = hover.Radius
					hover.unmark()

			else if (@target?)
				#Cursor just left a Cell
				if (@target.Owner == vires.Self)
					#That Cell was owned by the Player
					#Place source marker
					@target.mark()
				#Remove target marker
				@target = null
				@targetMarker.unlink()

		if (input.leftReleased && !@spectating)
			hover = @cellAt(input.cursor)
			if (hover?)
				#Send Movements
				#console.log @markers
				sources = []
				for id, marked of @markers
					if(marked.cell.ID != hover.ID)
						sources.push(marked.cell)
				#console.log "---"
				#console.log sources
				#console.log "-@-"
				#console.log hover
				#console.log "---"
				connection.sendMove(hover, sources)
			#Remove all markers
			@target = null
			@markers = { }
			gfx.material.marker.clear()
					

		#Perform Camera movement
		if(input.rightPressed)
			@cameraDrag = true
			@cameraStart = input.cursor
		if(input.rightReleased)
			@cameraDrag = false
		if(@cameraDrag)
			delta = vec2.subtract(vec2.create(), @cameraStart, input.cursor)
			vec2.add(gfx.camera.pos, gfx.camera.pos, delta)

		#Perform Camera zoom
		if(input.scroll != 0 && !@cameraDrag)
			prevZoom = gfx.camera.zoom
			if(input.scroll > 0)
				gfx.camera.zoom *= 1 + input.scroll * settings.zoomSpeed
				if(gfx.camera.zoom>settings.maxZoom)
					gfx.camera.zoom = settings.maxZoom
			else
				gfx.camera.zoom /= 1 - input.scroll * settings.zoomSpeed
				if(gfx.camera.zoom<settings.minZoom)
					gfx.camera.zoom = settings.minZoom

			#Zoom towards cursor position
			zoomFactor = gfx.camera.zoom/prevZoom
			offset = vec2.subtract(vec2.create(), input.cursor, gfx.camera.pos)
			lerp = 1 - 1/zoomFactor
			vec2.scale(offset, offset, lerp)
			vec2.add(gfx.camera.pos, gfx.camera.pos, offset)

		#Trap Camera in Field confinds
		vec2.max(gfx.camera.pos, gfx.camera.pos, vec2.create())
		vec2.min(gfx.camera.pos, gfx.camera.pos, @fieldSize)
		return

	digestTraffic: ->
		Msg = connection.messages.pop()
		while(Msg?)
			data = Msg.Data
			switch Msg.Type
				when "Movement"
					@movements[data.ID] = new Movement(data)
					#console.log "Movement #{data.ID}"
					#console.log @movements[data.ID]
				when "Replication"
					for update in data
						@cells[update.ID].update(update.Stationed)
				when "Conflict"
					@movements[data.Movement].kill()
					delete @movements[data.Movement]
					@cells[data.Cell.ID].update(data.Cell.Stationed)
					@cells[data.Cell.ID].switchOwner(@players[data.Cell.Owner])
				when "Collision"
					A = @movements[data.A.ID]
					B = @movements[data.B.ID]
					#console.log "Collision #{data.A.ID} #{data.B.ID}"
					#console.log A
					#console.log B
					if(data.A.Moving > 0)
						A.update(data.A)
					else
						A.kill()
						delete @movements[A.ID]
					if(data.B.Moving > 0)
						B.update(data.B)
					else
						B.kill()
						delete @movements[B.ID]
				when "EliminatedPlayer"
					if(vires.Self == data)
						@spectating = true
					@killPlayer(data)
				when "Winner"
					vires.load("lobby", @players[data].color)
				else
					connection.defaultDigest(Msg)
			Msg = connection.messages.pop()
		return

	animate: ->
		time = vires.time
		for k, mov of @movements
			mov.move(time)
		
		return


	clearMarkers: ->
		gfx.material.marker.clear()
		return

	killPlayer: (ID)->
		@players[ID].kill()
		for cell in @cells
			if (cell.Owner == ID)
				cell.Owner = 0
		for move in @movements
			if (move.Owner == ID)
				move.kill()
				delete @movements[move.ID]
		return
		
	cellAt: (pos)->
		for cell in @lookup
			if(vec2.distance(input.cursor, cell.Pos) <= cell.Radius)
				return cell
		return null

	#returns the Cell at the given position,
	# null if there is none
	cellAtFast: (pos)->
		targetMin = pos[x]-@maxCellSize
		targetMax = pos[x]+@maxCellSize
		lowBound = @lookup.length - 1
		highBound = 0
		min = 0
		max = @lookup.length - 1
		searching = true
		while (max-min >= 0)
			head = min + Math.ceil((max-min) / 2)
			current = @lookup[head].Pos[x]
			
			if(current - targetMin < 0)
				#below target zone
				min = head + 1
				if(highBound < head)
					highBound = head
			else 
				max = head - 1
				if(targetMax - current < 0)
					#above target zone
					if(lowBound > head)
						lowBound = head
				else
					#in target zone
					if(highBound < head)
						highBound = head
					if(lowBound < head)
						lowBound = head
		min = Math.max(highBound, lowBound) + 1
		max = @lookup.length - 1
		while (max-min >= 0)
			head = min + Math.ceil((max-min) / 2)
			current = @lookup[head].Pos[x]
			
			if(targetMax - current < 0)
				#above target zone
				max = head - 1
				if(highBound < head)
					highBound = head
			else 
				min = head + 1
				if !(current - targetMin < 0)
					#in target zone
					if(highBound < head)
						highBound = head
		#Check each cell that is between lowBound and highBound
		if(highBound - lowBound >= 0)
			for i in [lowBound..highBound]
				if(vec2.distance(input.cursor, @lookup[i].Pos) <= @lookup[i].Radius)
					return @lookup[i]
		return null


vires.states.loading = 
	animation: []

	load: ->
		vec2.set(gfx.camera.pos, 0, 0)
		gfx.camera.zoom = 5

		color = gfx.makeColor(2)
		mesh = gfx.mesh.round
		material = gfx.material.loading
		@animation = new Array(10)
		for i in [0...10]
			@animation[i] = vec2.create()
			segment = new Primitive(@animation[i], mesh, material, i)
			segment.scale = 0.5 + i*0.05
			segment.color = vec4.clone(color)
			segment.color[3] = 0.1 + i*0.1
		return

	unload: ->
		gfx.material.loading.clear()
		return

	digestInput: ->
		return

	digestTraffic: ->
		#Connection is not established during this state
		return

	animate: ->
		for i in [0...9]
			vec2.copy(@animation[i], @animation[i+1])
		distance = vec2.distance(input.cursor, @animation[9])
		if(distance != 0)
			step = Math.min(2, distance)
			step = step/distance
			vec2.lerp(@animation[9], @animation[9], input.cursor, step)
		return


vires.states.lobby = 
	animation: []

	load: (winner)->
		vec2.set(gfx.camera.pos, 0, 0)
		gfx.camera.zoom = 5

		color = null
		if(winner?)
			color = winner
		else
			color = gfx.makeColor(10)
		mesh = gfx.mesh.round
		material = gfx.material.loading
		@animation = new Array(10)
		for i in [0...10]
			@animation[i] = vec2.create()
			segment = new Primitive(@animation[i], mesh, material, i)
			segment.color = vec4.clone(color)
			segment.color[3] = 0.1 + i*0.1
		return

	unload: ->
		gfx.material.loading.clear()
		return

	digestInput: ->
		return

	digestTraffic: ->
		Msg = connection.messages.pop()
		while(Msg?)
			connection.defaultDigest(Msg)
			Msg = connection.messages.pop()
		return

	animate: vires.states.loading.animate


vires.states.noConnection = 

	load: ->
		return

	unload: ->
		return

	digestInput: ->
		return

	digestTraffic: ->
		#Connection has been cut off
		return

	animate: ->
		return

#Only used for debug purposes
vires.states.debug = 

	load: ->
		return

	unload: ->
		return

	digestInput: ->
		return

	digestTraffic: ->
		return

	animate: ->
		return

#Takes an in-game coordinate and calculates
# the equivalent screen-space coordinate
convertGameCoords = (pos)->
	out = vec2.subtract(vec2.create(), pos, gfx.camera.pos)
	vec2.scale(out, out, gfx.camera.zoom)
	out[1] = -out[1]
	out[0] += gfx.width/2
	out[1] += gfx.height/2
	return out

#Takes a coordinate in screen-space
# and calculates the equivalent in-game coordinate
convertMouseCoords = (pos)->
	out = vec2.fromValues(pos[0] - gfx.width/2, gfx.height/2 - pos[1])
	vec2.scale(out, out, 1/gfx.camera.zoom)
	vec2.add(out, out, gfx.camera.pos)
	return out

#A seeded random number generator
# used to create the same sequence of
# pseudo random numbers on all clients
class Random
	seed: 0

	constructor: (@seed)->
		return

	#Produces a number between 0 and 1
	next: =>
		@seed = (@seed * 9301 + 49297) % 233280
		return @seed / 233280

	#Produced a number inside the given range
	nextIn: (min, max)=>
		@seed = (@seed * 9301 + 49297) % 233280
		rnd = @seed / 233280

		return min + rnd * (max - min)

	#Shuffles all entries of an array
	shuffle: (arr) ->
		i = arr.length
		while --i > 0
			j = ~~(@next() * (i + 1))
			t = arr[j]
			arr[j] = arr[i]
			arr[i] = t
		return arr


class Player
	#Unique identifier sent by the server
	ID: 0
	#Color of this Player's Cells
	#Never swap these for other colors.
	# modify these colors instead. (vec4.set)
	color: null
	colorLight: null
	colorDark: null

	alive: true
	
	constructor: (@ID)->
		@alive = true
		color: gfx.color[0]
		colorDark: vec4.lerp(vec4.create(), @color, gfx.black, settings.factorDark)
		colorLight: vec4.lerp(vec4.create(), @color, gfx.white, settings.factorLight)
		return

	swapColor: (color)->
		vec4.copy(@color, color)
		vec4.lerp(@colorLight, @color, gfx.white, settings.factorLight)
		vec4.lerp(@colorDark, @color, gfx.black, settings.factorDark)

	kill: ->
		vec4.copy(@color, gfx.color[0])
		@alive = false


class Cell 
	#Unique identifier sent by the server
	ID: 0
	
	Owner: 0
	Pos: vec2.fromValues(0, 0)
	Radius: 1
	Capacity: 1
	Stationed: 0
	#This oject's graphical representation
	body: null
	gauge: null
	antigauge: null
	marker: null

	#Constructs a Cell out of data, which
	# was received from the server
	constructor: (Data)->
		@ID = Data.ID
		@Pos = vec2.fromValues(Data.Body.Location.X, Data.Body.Location.Y)
		@Radius = Data.Body.Radius
		@Capacity = Data.Capacity

		@body = new Primitive(@Pos, gfx.mesh.round, gfx.material.cell)
		@gauge = new Primitive(@Pos, gfx.mesh.round, gfx.material.cell)
		@antigauge = new Primitive(@Pos, gfx.mesh.round, gfx.material.cell)
		@marker = new Primitive(@Pos, gfx.mesh.mark, gfx.material.marker)
		@switchHeight(settings.indexNeutral)
		@body.color = gfx.color[0]
		@gauge.color = gfx.black
		@antigauge.color = gfx.color[0]
		@body.scale = @Radius
		@marker.scale = @Radius
		@gauge.scale = 0
		@antigauge.scale = 0
		return

	switchOwner: (owner)->
		@Owner = owner.ID
		if (@Owner == vires.Self)
			@switchHeight(settings.indexSelf)
		else if (@Owner == 0)
			@switchHeight(settings.indexNeutral)
		else
			@switchHeight(settings.indexOther)
		@body.color = owner.color
		@antigauge.color = owner.color
		return

	switchHeight: (height)->
		@body.height = height
		@gauge.height = height + 1
		@antigauge.height = height + 2
		@marker.height = height + 5

	update: (stationed)->
		@Stationed = stationed
		fullnes = stationed /  @Capacity
		trailing = Math.max( fullnes-settings.gauge, 0)
		fullnes *= @Radius
		trailing *= @Radius
		@gauge.scale = fullnes
		@antigauge.scale = trailing
		return

	mark: ->
		@mark.link()
		return

	unmark: ->
		@mark.unlink()
		return


class Movement
	#Unique identifier sent by the server
	ID: 0
	#ID of the Play, who owns this Cell
	Owner: 0
	Moving: 0
	#This Movements origin
	O: vec2.fromValues(0, 0)
	Radius: 1
	#This Movements speed vector
	V: vec2.fromValues(0, 0)
	#Moment in in-game time at which
	# this Movement started
	birth: 0
	pos: vec2.fromValues(0, 0)
	#This oject's graphical representation
	body: null

	#Constructs a Movement out of data, which
	# was received from the server
	constructor: (Data)->
		@ID = Data.ID
		@Owner = Data.Owner
		@Moving = Data.Moving
		@O = vec2.fromValues(Data.Body.Location.X, Data.Body.Location.Y)
		@Radius = Data.Body.Radius
		@V = vec2.fromValues(Data.Direction.X, Data.Direction.Y)
		@birth = vires.time
		@pos = vec2.clone(@O)

		@body = new Primitive(@pos, gfx.mesh.round, gfx.material.movement)
		if (@Owner == vires.Self)
			@body.height = settings.indexSelf + settings.offsetMovement
		else if (@Owner == 0)
			@body.height = settings.indexNeutral + settings.offsetMovement
		else
			@body.height = settings.indexOther + settings.offsetMovement
		@body.scale = @Radius
		@body.color = vires.states.match.players[@Owner].color
		return

	#Updates the position of this Movement
	move: (now)->
		vec2.scaleAndAdd(@pos, @O, @V, now-@birth)
		return

	#Stops this Movement from being displayed
	kill: ->
		@body.unlink()
		return

	#Modifies this movement after a Collision
	# was received from the server
	update: (Data)->
		#console.log "updating #{@ID}"
		@Moving = Data.Moving
		#console.log "O: "
		#console.log @O[0], @O[1]
		vec2.set(@O, Data.Body.Location.X, Data.Body.Location.Y)
		#console.log @O[0], @O[1]
		@Radius = Data.Body.Radius
		#console.log "V: "
		#console.log @V[0], @V[1]
		vec2.set(@V, Data.Direction.X, Data.Direction.Y)
		#console.log @V[0], @V[1]
		#console.log "birth: #{@birth}"
		@birth = vires.time
		#console.log @birth
		#console.log "pos:"
		#console.log @pos[0], @pos[1]
		vec2.copy(@pos, @O)
		#console.log @pos[0], @pos[1]
		#console.log "--updated!"

		@body.scale = @Radius
		return
