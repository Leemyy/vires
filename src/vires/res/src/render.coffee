#All resouces, that sould be loaded into Graphics memory
# This is used every time, the GLContext gets initialized
gfx = 
	width: 300
	height: 150
	#All graphics ressources need to be appended here to be loaded
	# into Graphics memory on startup and after loss of context.
	resources: new Array(0)
	shader: {}
	material: {}
	mesh: {}
	texture: {}
	color: new Array(0)

	boundMesh: ""
	camera: null
	matVP: mat4.create()

	init: ->
		GL.enable(GL.DEPTH_TEST)
		GL.depthFunc(GL.GREATER)
		GL.clearColor(1.0, 1.0, 1.0, 1.0)
		GL.clearDepth(-1.0)
		
		#Load All resources into Graphics memory
		for res in gfx.resources
			res.load()

		gfx.camera = new OrthoCam(vec3.fromValues(0, 0, 1000))
		return

	addShader: (nShader)->
		@resources.push(nShader)
		@shader[nShader.name] = nShader
		return

	addMaterial: (nMaterial)->
		#@resources.push(nMaterial)
		@material[nMaterial.name] = nMaterial
		return

	addMesh: (nMesh)->
		@resources.push(nMesh)
		@mesh[nMesh.name] = nMesh
		return

	addTexture: (nTexture)->
		@resources.push(nTexture)
		@texture[nTexture.name] = nTexture
		return

	drawScene: ->
		GL.viewport(0.0, 0.0, GL.drawingBufferWidth, GL.drawingBufferHeight)
		GL.clear(GL.COLOR_BUFFER_BIT | GL.DEPTH_BUFFER_BIT)
		gfx.matVP = mat4.multiply(mat4.create(), gfx.camera.projMatrix(gfx.width, gfx.height), gfx.camera.viewMatrix())

		for shaName, shader of gfx.shader
			shader.use()
			shader.drawMaterials()

		return

	makeColor: (index)->
		return vec4.clone(@color[index])


#A wrapper for WebGL shaderprograms
class Program
	name: "?"

	#A handle to the WebGL program
	id: null

	source:
		vert: "" 
		frag: ""

	shader:
		vert: null 
		frag: null 

	err: 
		any: no 
		vert: "" 
		frag: "" 
		prog: "" 

	attribute: null

	uniform: null

	materials: null

	constructor: (@name, attributes, uniforms, srcVert, srcFrag, drawFunc ) ->

		@source = {}
		@source.vert = srcVert
		@source.frag = srcFrag

		@shader = {}
		@err = {}

		@attribute = {}
		for i in [0...attributes.length]
			@attribute[attributes[i]] = null
		
		@uniform = {}
		for i in [0...uniforms.length]
			@uniform[uniforms[i]] = null

		@materials = new Array(0)

		@draw = drawFunc

		gfx.addShader(this)
		return

	#placeholder for implementation specific drawing
	draw: ->
		console.err("Shader #{this.name} has no drawing function")
		return

	drawMaterials: ->
		for material in @materials
			@draw(material)
			# ...
		return

	addMaterial: (material)->
		@materials.push(material)

	#Load shaders into graphics memory, compile them
	# and retrieve the attribute and uniform positions
	load: ->
		@shader.vert = GL.createShader(GL.VERTEX_SHADER)
		GL.shaderSource(@shader.vert, @source.vert)
		GL.compileShader(@shader.vert)
		if (!GL.getShaderParameter(@shader.vert, GL.COMPILE_STATUS))
			@err.any = yes
			@err.vert = GL.getShaderInfoLog(@shader.vert)	

		@shader.frag = GL.createShader(GL.FRAGMENT_SHADER)
		GL.shaderSource(@shader.frag, @source.frag)
		GL.compileShader(@shader.frag)
		if (!GL.getShaderParameter(@shader.frag, GL.COMPILE_STATUS))
			@err.any = yes
			@err.frag = GL.getShaderInfoLog(@shader.frag)

		@id = GL.createProgram()
		GL.attachShader(@id, @shader.vert)
		GL.attachShader(@id, @shader.frag)
		GL.linkProgram(@id)

		if (!GL.getProgramParameter(@id, GL.LINK_STATUS)) 
			@err.any = yes
			@err.prog = "Program linking failed"
		else
			GL.useProgram(@id)

			for key, val of @attribute
				@attribute[key] = GL.getAttribLocation(@id, key)
				GL.enableVertexAttribArray(@attribute[key])
			
			for key, val of @uniform
				@uniform[key] = GL.getUniformLocation(@id, key)

			GL.useProgram(null)

		if @err.any
			console.err("Error compiling shader: #{[key, value] for key, value of @err}" )
		return

	use: ->
		GL.useProgram(@id)
		return @id


class Mesh
	name: "?"
	data: null
	faces: null

	buffer: null
	elements: null

	stride: 0
	vertexCount: 0
	vertexOffset: 0

	constructor: (@name, @data, @faces, @stride, @vertexCount, @vertexOffset) ->
		for i in [0...@faces.length]
			@faces[i] -= 1
		
		gfx.addMesh(this)
		return

	load: ->
		@buffer = GL.createBuffer()
		GL.bindBuffer(GL.ARRAY_BUFFER, @buffer)
		GL.bufferData(GL.ARRAY_BUFFER, new Float32Array(@data), GL.STATIC_DRAW)

		@elements = GL.createBuffer()
		GL.bindBuffer(GL.ELEMENT_ARRAY_BUFFER, @elements)
		GL.bufferData(GL.ELEMENT_ARRAY_BUFFER, new Uint16Array(@faces), GL.STATIC_DRAW)
		return

	bind: ->
		if(gfx.boundMesh!=@name)
			gfx.boundMesh = @name
			GL.bindBuffer(GL.ARRAY_BUFFER, @buffer)
			GL.bindBuffer(GL.ELEMENT_ARRAY_BUFFER, @elements)
		return


class Material
	name: "?"
	instances: null

	constructor: (@name, shader)->
		@instances = new Array(0)
		gfx.addMaterial(this)
		shader.addMaterial(this)
		return

	register: (primitive)->
		primitive.index = @instances.length
		@instances[primitive.index] = primitive
		return

	remove: (primitive)->
		if (primitive.index == -1)
			return
		last = @instances.length - 1
		if (primitive.index == last)
			@instances.pop()
		else
			top = @instances.pop()
			top.index = primitive.index
			@instances[primitive.index] = top
		primitive.index = -1
		return

	clear: ->
		for primitive in @instances
			primitive.index = -1
		@instances = new Array(0)


class Primitive
	pos: null
	height: 0
	scale: 1
	mesh: null
	tex: null
	color: vec4.fromValues(0, 0, 0, 1)
	material: null
	index: -1

	constructor: (@pos, @mesh, @material, height, scale)->
		if(height?)
			@height = height
		if(scale?)
			@scale = scale
		@link()
		return

	modelMatrix: ->
		model = mat4.create()
		model = mat4.translate(model, model, vec3.fromValues(@pos[0], @pos[1], @height))

	link: ->
		if(@index < 0)
			@material.register(this)
		return

	unlink: ->
		@material.remove(this)
		return


class OrthoCam
	pos: vec3.fromValues(0, 0, 1) 
	look: vec3.fromValues(0, 0, -1) 
	up: vec3.fromValues(0, 1, 0) 

	near: 0
	far: 10000
	zoom: 1

	constructor: (cPos, cLook, cUp) ->
		if(cPos?)
			@pos = cPos
		else
			@pos = vec3.clone(@pos)

		if(cLook?)
			@look = cLook
		else
			@look = vec3.clone(@look)

		if(cUp)
			@up = cUp
		else
			@up = vec3.clone(@up)
		return

	viewMatrix:  ->
		antiPos = vec3.create()
		vec3.negate(antiPos, @pos)
		translate = mat4.create()
		mat4.fromTranslation(translate, antiPos)

		xLocal = vec3.create()
		yLocal = vec3.create()
		zLocal = vec3.create()
		view = mat4.create()
		vec3.cross(xLocal, @look, @up)
		vec3.normalize(xLocal, xLocal)
		vec3.cross(yLocal, xLocal, @look)
		vec3.normalize(yLocal, yLocal)
		vec3.negate(zLocal, @look)
		view[0] = xLocal[0]
		view[1] = yLocal[0]
		view[2] = zLocal[0]
		view[4] = xLocal[1]
		view[5] = yLocal[1]
		view[6] = zLocal[1]
		view[8] = xLocal[2]
		view[9] = yLocal[2]
		view[10] = zLocal[2]

		mat4.multiply(view, view, translate)
		return view

	projMatrix: (width, height) ->
		width /= 2*@zoom
		height /= 2*@zoom
		proj = mat4.create()
		mat4.ortho(proj, -width, width, -height, height, @far, @near)
		return proj

	viewRange: (@near, @far) ->
		return

	lookAt: (vx, y, z) ->
		if typeof vx == "number"
			vx = vec3.fromValues(vx, y, z)
		vec3.subtract(@look, vx, @pos)
		return vec3.normalize(@look, @look)

	lookIn: (vx, y, z) ->
		if typeof vx == "number"
			vx = vec3.fromValues(vx, y, z)
		return vec3.normalize(@look, vx)

	moveTo: (vx, y, z) ->
		if typeof vx == "number"
			vx = vec3.fromValues(vx, y, z)
		return @pos = vx

	moveBy: (vx, y, z) ->
		if typeof vx == "number"
			vx = vec3.fromValues(vx, y, z)
		return vec3.add(@pos, @pos, vx)

	orientUp: (vx, y, z) ->
		if typeof vx == "number"
			vx = vec3.fromValues(vx, y, z)
		return vec3.normalize(@up, vx)

	orientGravity: (vx, y, z) ->
		if typeof vx == "number"
			vx = vec3.fromValues(vx, y, z)
		vec3.subtract(@up, @pos, vx)
		return vec3.normalize(@up, @up)


class Camera extends OrthoCam
	lens: glMatrix.toRadian(90)
	fov: @lens

	projMatrix: (width, height) ->
		ratio = width/height
		proj = mat4.create()
		return mat4.perspective(proj, @fov, ratio, @near, @far)

	setZoom: (@zoom) ->
		return @fov = Math.atan(Math.tan(@lens)/@zoom)

	setLens: (@lens) ->
		@fov = @lens
		return




	
