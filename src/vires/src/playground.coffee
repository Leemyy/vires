
connection=
  version: "0.1"
  url: "ws://localhost"
  messages: [1, 2, 3, 4]
  socket: null



  move: (target, attackers)->
    attacks = new Array(attackers.length)
    for i in [0...attackers.length]
      attacks[i] =
        Source: attackers[i].ID
        Dest: target.ID
      # ...

  send: (type, payload)->
    packet = 
      Type: type
      Version: @version
      Data: payload
    data = JSON.stringify(packet)
    socket.send(JSON.stringify(packet))

for i in [1..0]
  console.log("step #{i}")
  # ...


###Convert Colors
lit= """
{ "set" :[
[240,163,255],
[0,117,220],
[153,63,0],
[76,0,92],
[25,25,25],
[0,92,49],
[43,206,72],
[255,204,153],
[128,128,128],
[148,255,181],
[143,124,0],
[157,204,0],
[194,0,136],
[0,51,128],
[255,164,5],
[255,168,187],
[66,102,0],
[255,0,16],
[94,241,242],
[0,153,143],
[224,255,102],
[116,10,255],
[153,0,0],
[255,255,128],
[255,255,0],
[255,80,5]
]
}
"""
obj = {}
try
  obj = JSON.parse(lit)
catch e
  console.error(e)

text = for raw in obj.set
 line= "new Material(\"\", vec3.fromValues(#{raw[0]/255}, #{raw[1]/255}, #{raw[2]/255}, 1))"
 console.log(line)
###


###
http://www.webglacademy.com/
https://www.khronos.org/webgl/wiki/Tutorial
http://learningwebgl.com/blog/?p=28

,
  
  multiply: function (out, a, b) {
		var a00 = a[0], a01 = a[1], a02 = a[2], a03 = a[3],
			a10 = a[4], a11 = a[5], a12 = a[6], a13 = a[7],
			a20 = a[8], a21 = a[9], a22 = a[10], a23 = a[11],
			a30 = a[12], a31 = a[13], a32 = a[14], a33 = a[15];

		// Cache only the current line of the second matrix
		var b0  = b[0], b1 = b[1], b2 = b[2], b3 = b[3];
		out[0] = b0*a00 + b1*a10 + b2*a20 + b3*a30;
		out[1] = b0*a01 + b1*a11 + b2*a21 + b3*a31;
		out[2] = b0*a02 + b1*a12 + b2*a22 + b3*a32;
		out[3] = b0*a03 + b1*a13 + b2*a23 + b3*a33;

		b0 = b[4]; b1 = b[5]; b2 = b[6]; b3 = b[7];
		out[4] = b0*a00 + b1*a10 + b2*a20 + b3*a30;
		out[5] = b0*a01 + b1*a11 + b2*a21 + b3*a31;
        out[6] = b0*a02 + b1*a12 + b2*a22 + b3*a32;
        out[7] = b0*a03 + b1*a13 + b2*a23 + b3*a33;

        b0 = b[8]; b1 = b[9]; b2 = b[10]; b3 = b[11];
        out[8] = b0*a00 + b1*a10 + b2*a20 + b3*a30;
        out[9] = b0*a01 + b1*a11 + b2*a21 + b3*a31;
        out[10] = b0*a02 + b1*a12 + b2*a22 + b3*a32;
        out[11] = b0*a03 + b1*a13 + b2*a23 + b3*a33;

		b0 = b[12]; b1 = b[13]; b2 = b[14]; b3 = b[15];
        out[12] = b0*a00 + b1*a10 + b2*a20 + b3*a30;
        out[13] = b0*a01 + b1*a11 + b2*a21 + b3*a31;
        out[14] = b0*a02 + b1*a12 + b2*a22 + b3*a32;
        out[15] = b0*a03 + b1*a13 + b2*a23 + b3*a33;
        return out;
	}



  GL.enable(GL.DEPTH_TEST);
  GL.depthFunc(GL.LEQUAL);
  GL.clearColor(0.0, 0.0, 0.0, 0.0);
  GL.clearDepth(1.0);

  var time_old=0;
  var animate=function(time) {
    var dt=time-time_old;
    if (!drag) {
      dX*=AMORTIZATION, dY*=AMORTIZATION;
      THETA+=dX, PHI+=dY;
    }
    LIBS.set_I4(MOVEMATRIX);
    LIBS.rotateY(MOVEMATRIX, THETA);
    LIBS.rotateX(MOVEMATRIX, PHI);
	LIBS.rotateY(ROTATION, AMORTIZATION/32);
    LIBS.multiply(MOVEMATRIX2, ROTATION, TRANSLATION);
    LIBS.rotateY(MOVEMATRIX2, count*AMORTIZATION/-16);
    count += 1;
    
    time_old=time;

    GL.viewport(0.0, 0.0, CANVAS.width, CANVAS.height);
    GL.clear(GL.COLOR_BUFFER_BIT | GL.DEPTH_BUFFER_BIT);
    GL.uniformMatrix4fv(_Pmatrix, false, PROJMATRIX);
    GL.uniformMatrix4fv(_Vmatrix, false, VIEWMATRIX);
    GL.uniformMatrix4fv(_Mmatrix, false, MOVEMATRIX);
    GL.bindBuffer(GL.ARRAY_BUFFER, CUBE_VERTEX);
    GL.vertexAttribPointer(_position, 3, GL.FLOAT, false,4*(3+3),0) ;
    GL.vertexAttribPointer(_color, 3, GL.FLOAT, false,4*(3+3),3*4) ;
    GL.bindBuffer(GL.ELEMENT_ARRAY_BUFFER, CUBE_FACES);
    GL.drawElements(GL.TRIANGLES, 6*2*3, GL.UNSIGNED_SHORT, 0);

    GL.uniformMatrix4fv(_Mmatrix, false, MOVEMATRIX2);


    GL.drawElements(GL.TRIANGLES, 6*2*3, GL.UNSIGNED_SHORT, 0);
    GL.flush();

    window.requestAnimationFrame(animate);
  };

  ###