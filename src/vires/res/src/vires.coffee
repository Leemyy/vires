#Variables to more intuitively refer to Vector components
x = r = 0
y = g = 1
z = b = 2
w = a = 3

#Stores all relevant HTML elements
html =
	body: null
	viewport: null
	overlay: null
	menu: {}

#Stores all relevant inputs, so the according actions 
# can be executed in the next frame
input =
	focus: true
	right: no
	rightPressed: no
	rightReleased: no
	middle: no
	middlePressed: no
	middleReleased: no
	left: no
	leftPressed: no
	leftReleased: no
	scroll: 0
	x:0
	y:0
	dx:0
	dy:0
	cursor: vec2.fromValues(0, 0)
	delta: vec2.fromValues(0, 0)

	#Directly modified by input listeners
	# (may change during a sigle frame)
	next:
		focus: true
		right: no
		middle: no
		left: no
		scroll: 0
		x:0
		y:0


#The WebGLRenderingContext
GL = {}



#Initialization method.
#Gets called once the HTML body is fully loaded
initialize = ->
	#Store important HTML elements for easy access
	html.overlay = document.getElementById("overlay")
	html.viewport = document.getElementById("viewport")
	html.body = document.body
	#Hide the menu overlay
	showMenu(false)

	#Add event listeners
	html.viewport.oncontextmenu = suppressEvent
	html.viewport.onmousedown = mousePressed
	html.viewport.onmouseup = mouseReleased
	html.viewport.onwheel = mouseWheel
	html.viewport.onmousemove = mouseMoved
	html.viewport.onmouseleave = mouseLeft
	html.viewport.onmouseenter = mouseEntered

	#Attempt to create the RenderingContext
	if 	initializeGL()
		html.body.onresize = resizeGL
		resizeGL(null)
		#Load resources and start the game
		prepareLoop()
	else 
		console.log("Error creating WebGL context!")
		#Inform the user about the WebGL error
		# and point him to https://get.webgl.org/
	return

#Fetches the WebGLRenderingContext for the viewport
# and sets some WebGL variables
initializeGL = ->
	if (!window.WebGLRenderingContext)
		# Browser does not support WebGL
		return false
	else
		GL = html.viewport.getContext("experimental-webgl", {antialias: true})

		if (!GL)
			# WebGL could not be initialized
			return false

		GL.enable(GL.DEPTH_TEST)
		GL.depthFunc(GL.GREATER)
		GL.clearColor(1.0, 1.0, 1.0, 1.0)
		GL.clearDepth(-1.0)
	return true
	

resizeGL = (event) ->
	#Change the size of the canvas to tell WebGL
	# to change the resolution of the rendered image
	gfx.height = html.viewport.height = html.viewport.clientHeight*devicePixelRatio
	gfx.width = html.viewport.width = html.viewport.clientWidth*devicePixelRatio
	return

#Shows and hides the menu (unused)
showMenu = (doShow) ->
	if doShow
		html.overlay.style.display = "block"
		html.overlay.focus()
	else
		html.overlay.style.display = "none"
		html.overlay.blur()
	return

#Suppresses the default behaviour and
# further propagation of an event.
#Used to prevent the popup menu from opening
suppressEvent = (event) ->
	event.preventDefault()
	event.stopPropagation()
	return


mousePressed = (event) ->
	switch event.button
		when 0
			input.next.left = yes
		when 1
			input.next.middle = yes
		when 2
			input.next.right = yes
	return

mouseReleased = (event) ->
	switch event.button
		when 0
			input.next.left = no
		when 1
			input.next.middle = no
		when 2
			input.next.right = no
	return

mouseWheel = (event) ->
	input.next.scroll -= event.deltaY
	return

mouseMoved = (event) ->
	input.next.x = event.clientX*devicePixelRatio
	input.next.y = event.clientY*devicePixelRatio
	return

mouseLeft = (event) ->
	input.next.focus = false
	return

mouseEntered = (event) ->
	input.next.focus = true
	return

#Copies the current input state to other variables
# to guarantee consistent values throughout each frame
#Also calculates delta values and button presses/releases
nextInput = ->
	justFocused = input.next.focus && !input.focus
	justBlurred = !input.next.focus && input.focus
	input.focus = input.next.focus
	#Only calculate deltas, if the game is focused
	if(input.focus)
		if(justFocused)
			input.dx = 0
			input.dy = 0	
		else
			input.dx = input.next.x - input.x
			input.dy = input.next.y - input.y


		input.rightPressed = input.next.right && !input.right
		input.middlePressed = input.next.middle && !input.middle
		input.leftPressed = input.next.left && !input.left

		input.rightReleased = !input.next.right && input.right
		input.middleReleased = !input.next.middle && input.middle
		input.leftReleased = !input.next.left && input.left
	else
		input.dx = 0
		input.dy = 0

		input.rightPressed = false
		input.middlePressed = false
		input.leftPressed = false

		#Release buttons, if focus was just lost
		if (justBlurred)
			input.rightReleased = true
			input.middleReleased = true
			input.leftReleased = true
		else
			input.rightReleased = false
			input.middleReleased = false
			input.leftReleased = false

	input.right = input.next.right
	input.middle = input.next.middle
	input.left = input.next.left
	input.scroll = input.next.scroll
	input.x = input.next.x
	input.y = input.next.y

	#Calculate in-game cursor movement
	cursor = vec2.fromValues(input.x, input.y)
	cursor = convertMouseCoords(cursor)
	if(justFocused || !input.focus)
		vec2.set(input.delta, 0, 0)
	else
		vec2.subtract(input.delta, cursor, input.cursor)
	input.cursor = cursor

	input.next.scroll = 0
	return

#Loads resources and starts the main game loop
prepareLoop = ->
	gfx.init()
	vires.init()
	connection.init()

	#load debug information for testing
	#loadDebug()


	requestAnimationFrame(gameLoop)
	return

#The main game loop
gameLoop = (timeNow)->
	vires.time = timeNow/1000
	vires.delta = vires.time - vires.timePrev

	#Perform state change
	if(vires.next.state?)
		vires.active.unload()
		vires.active = vires.next.state
		vires.next.state = null
		vires.active.load(vires.next.data)

	#React to new user input
	nextInput()
	vires.active.digestInput()

	#Send input to server and
	# execute recieved commands
	vires.active.digestTraffic()

	#Update entity positions
	vires.active.animate()

	#Render the current state
	gfx.drawScene()

	vires.timePrev = timeNow
	#Request Browser to call this function
	# again on the next redraw
	requestAnimationFrame(gameLoop)
	return
	

#only for testing purposes
loadDebug = ->
	vires.load("match", packets.field.Data)
	return
