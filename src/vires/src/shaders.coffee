
new Program( "basic",
#Attributes
	["a_VertexPosition"],
#Uniforms
	["u_MVPMatrix"],
#Vertex Shader
"""

attribute vec3 a_VertexPosition;

uniform mat4 u_MVPMatrix;

void main(void) {
	gl_Position = u_MVPMatrix * vec4(a_VertexPosition, 1.0);
}

""", 
#Fragment Shader
"""

precision mediump float;

void main(void) {
	gl_FragColor = vec4(0.0, 0.0, 0.0, 1.0);
}

""",
#Draw function
(material) ->
	
	for primitive in material.instances
		GL.uniformMatrix4fv(@uniform.u_MVPMatrix, false, mat4.multiply(mat4.create(), gfx.matVP, primitive.modelMatrix()))
		#GL.uniformMatrix4fv(@uniform.u_MVPMatrix, false, mat4.fromTranslation(mat4.create(), primitive.pos))
		GL.bindBuffer(GL.ARRAY_BUFFER, primitive.mesh.buffer)
		GL.vertexAttribPointer(@attribute.a_VertexPosition, 3, GL.FLOAT, false,primitive.mesh.stride,primitive.mesh.vertexOffset)
		GL.bindBuffer(GL.ELEMENT_ARRAY_BUFFER, primitive.mesh.elements)
		GL.drawElements(GL.TRIANGLES, primitive.mesh.vertexCount, GL.UNSIGNED_SHORT, 0)
	
)

new Program( "color",
#Attributes
	["a_VertexPosition"],
#Uniforms
	["u_MVPMatrix", "u_Scale", "u_Color"],
#Vertex Shader
"""

attribute vec3 a_VertexPosition;

uniform mat4 u_MVPMatrix;
uniform vec3 u_Scale;

void main(void) {
	gl_Position = u_MVPMatrix * vec4(a_VertexPosition * u_Scale, 1.0);
}

""", 
#Fragment Shader
"""

precision mediump float;

uniform vec4 u_Color;

void main(void) {
	gl_FragColor = u_Color;
}

""",
#Draw function
(material) ->
	
	for primitive in material.instances
		mat = mat4.multiply(mat4.create(), gfx.matVP, primitive.modelMatrix())
		GL.uniformMatrix4fv(@uniform.u_MVPMatrix, false, mat4.multiply(mat4.create(), gfx.matVP, primitive.modelMatrix()))
		GL.uniform3f(@uniform.u_Scale, primitive.scale, primitive.scale, primitive.scale)
		GL.uniform4fv(@uniform.u_Color, primitive.color)
		#GL.uniformMatrix4fv(@uniform.u_MVPMatrix, false, mat4.fromTranslation(mat4.create(), primitive.pos))
		GL.bindBuffer(GL.ARRAY_BUFFER, primitive.mesh.buffer)
		GL.vertexAttribPointer(@attribute.a_VertexPosition, 3, GL.FLOAT, false,primitive.mesh.stride,primitive.mesh.vertexOffset)
		GL.bindBuffer(GL.ELEMENT_ARRAY_BUFFER, primitive.mesh.elements)
		GL.drawElements(GL.TRIANGLES, primitive.mesh.vertexCount, GL.UNSIGNED_SHORT, 0)
	
)

### Excerpt from WebGL documentation
void glVertexAttribPointer( GLuint index, size, GLenum type, false, stride, offset)
	size [1-4]: number of values per attribute. (size = 3 for a vec3)
	stride : number of values between attribute starts. (stride = 0 means tightly packed)
	offset : number of values before the first attribute. 

void uniformMatrix( GLuint index, bool transpose, float[] data)

void glDrawElements( GLenum mode, GLsizei count, GLenum type, GLvoid indices)
	mode : [POINTS, LINES or TRIANGLES]
	count : Number of Vertices
	type : Type of the values in indices [GL_UNSIGNED_BYTE or GL_UNSIGNED_SHORT]

###