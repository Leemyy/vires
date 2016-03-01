// Generated by CoffeeScript 1.10.0
var Camera, Material, Mesh, OrthoCam, Primitive, Program, gfx,
  extend = function(child, parent) { for (var key in parent) { if (hasProp.call(parent, key)) child[key] = parent[key]; } function ctor() { this.constructor = child; } ctor.prototype = parent.prototype; child.prototype = new ctor(); child.__super__ = parent.prototype; return child; },
  hasProp = {}.hasOwnProperty;

gfx = {
  width: 300,
  height: 150,
  resources: new Array(0),
  shader: {},
  material: {},
  mesh: {},
  texture: {},
  color: new Array(0),
  boundMesh: "",
  camera: null,
  matVP: mat4.create(),
  init: function() {
    var j, len, ref, res;
    ref = gfx.resources;
    for (j = 0, len = ref.length; j < len; j++) {
      res = ref[j];
      res.load();
    }
    gfx.camera = new OrthoCam(vec3.fromValues(0, 0, 1000));
  },
  addShader: function(nShader) {
    this.resources.push(nShader);
    this.shader[nShader.name] = nShader;
  },
  addMaterial: function(nMaterial) {
    this.material[nMaterial.name] = nMaterial;
  },
  addMesh: function(nMesh) {
    this.resources.push(nMesh);
    this.mesh[nMesh.name] = nMesh;
  },
  addTexture: function(nTexture) {
    this.resources.push(nTexture);
    this.texture[nTexture.name] = nTexture;
  },
  drawScene: function() {
    var ref, shaName, shader;
    GL.viewport(0.0, 0.0, GL.drawingBufferWidth, GL.drawingBufferHeight);
    GL.clear(GL.COLOR_BUFFER_BIT | GL.DEPTH_BUFFER_BIT);
    gfx.matVP = mat4.multiply(mat4.create(), gfx.camera.projMatrix(gfx.width, gfx.height), gfx.camera.viewMatrix());
    ref = gfx.shader;
    for (shaName in ref) {
      shader = ref[shaName];
      shader.use();
      shader.drawMaterials();
    }
  },
  makeColor: function(index) {
    return vec4.clone(this.color[index]);
  }
};

Program = (function() {
  Program.prototype.name = "?";

  Program.prototype.id = null;

  Program.prototype.source = {
    vert: "",
    frag: ""
  };

  Program.prototype.shader = {
    vert: null,
    frag: null
  };

  Program.prototype.err = {
    any: false,
    vert: "",
    frag: "",
    prog: ""
  };

  Program.prototype.attribute = null;

  Program.prototype.uniform = null;

  Program.prototype.materials = null;

  function Program(name, attributes, uniforms, srcVert, srcFrag, drawFunc) {
    var i, j, k, ref, ref1;
    this.name = name;
    this.source = {};
    this.source.vert = srcVert;
    this.source.frag = srcFrag;
    this.shader = {};
    this.err = {};
    this.attribute = {};
    for (i = j = 0, ref = attributes.length; 0 <= ref ? j < ref : j > ref; i = 0 <= ref ? ++j : --j) {
      this.attribute[attributes[i]] = null;
    }
    this.uniform = {};
    for (i = k = 0, ref1 = uniforms.length; 0 <= ref1 ? k < ref1 : k > ref1; i = 0 <= ref1 ? ++k : --k) {
      this.uniform[uniforms[i]] = null;
    }
    this.materials = new Array(0);
    this.draw = drawFunc;
    gfx.addShader(this);
    return;
  }

  Program.prototype.draw = function() {
    console.err("Shader " + this.name + " has no drawing function");
  };

  Program.prototype.drawMaterials = function() {
    var j, len, material, ref;
    ref = this.materials;
    for (j = 0, len = ref.length; j < len; j++) {
      material = ref[j];
      this.draw(material);
    }
  };

  Program.prototype.addMaterial = function(material) {
    return this.materials.push(material);
  };

  Program.prototype.load = function() {
    var key, ref, ref1, val, value;
    this.shader.vert = GL.createShader(GL.VERTEX_SHADER);
    GL.shaderSource(this.shader.vert, this.source.vert);
    GL.compileShader(this.shader.vert);
    if (!GL.getShaderParameter(this.shader.vert, GL.COMPILE_STATUS)) {
      this.err.any = true;
      this.err.vert = GL.getShaderInfoLog(this.shader.vert);
    }
    this.shader.frag = GL.createShader(GL.FRAGMENT_SHADER);
    GL.shaderSource(this.shader.frag, this.source.frag);
    GL.compileShader(this.shader.frag);
    if (!GL.getShaderParameter(this.shader.frag, GL.COMPILE_STATUS)) {
      this.err.any = true;
      this.err.frag = GL.getShaderInfoLog(this.shader.frag);
    }
    this.id = GL.createProgram();
    GL.attachShader(this.id, this.shader.vert);
    GL.attachShader(this.id, this.shader.frag);
    GL.linkProgram(this.id);
    if (!GL.getProgramParameter(this.id, GL.LINK_STATUS)) {
      this.err.any = true;
      this.err.prog = "Program linking failed";
    } else {
      GL.useProgram(this.id);
      ref = this.attribute;
      for (key in ref) {
        val = ref[key];
        this.attribute[key] = GL.getAttribLocation(this.id, key);
        GL.enableVertexAttribArray(this.attribute[key]);
      }
      ref1 = this.uniform;
      for (key in ref1) {
        val = ref1[key];
        this.uniform[key] = GL.getUniformLocation(this.id, key);
      }
      GL.useProgram(null);
    }
    if (this.err.any) {
      console.log("Error compiling shader: " + ((function() {
        var ref2, results;
        ref2 = this.err;
        results = [];
        for (key in ref2) {
          value = ref2[key];
          results.push([key, value]);
        }
        return results;
      }).call(this)));
    }
  };

  Program.prototype.use = function() {
    GL.useProgram(this.id);
    return this.id;
  };

  return Program;

})();

Mesh = (function() {
  Mesh.prototype.name = "?";

  Mesh.prototype.data = null;

  Mesh.prototype.faces = null;

  Mesh.prototype.buffer = null;

  Mesh.prototype.elements = null;

  Mesh.prototype.stride = 0;

  Mesh.prototype.vertexCount = 0;

  Mesh.prototype.vertexOffset = 0;

  function Mesh(name, data, faces, stride, vertexCount, vertexOffset) {
    var i, j, ref;
    this.name = name;
    this.data = data;
    this.faces = faces;
    this.stride = stride;
    this.vertexCount = vertexCount;
    this.vertexOffset = vertexOffset;
    for (i = j = 0, ref = this.faces.length; 0 <= ref ? j < ref : j > ref; i = 0 <= ref ? ++j : --j) {
      this.faces[i] -= 1;
    }
    gfx.addMesh(this);
    return;
  }

  Mesh.prototype.load = function() {
    this.buffer = GL.createBuffer();
    GL.bindBuffer(GL.ARRAY_BUFFER, this.buffer);
    GL.bufferData(GL.ARRAY_BUFFER, new Float32Array(this.data), GL.STATIC_DRAW);
    this.elements = GL.createBuffer();
    GL.bindBuffer(GL.ELEMENT_ARRAY_BUFFER, this.elements);
    GL.bufferData(GL.ELEMENT_ARRAY_BUFFER, new Uint16Array(this.faces), GL.STATIC_DRAW);
  };

  Mesh.prototype.bind = function() {
    if (gfx.boundMesh !== this.name) {
      gfx.boundMesh = this.name;
      GL.bindBuffer(GL.ARRAY_BUFFER, this.buffer);
      GL.bindBuffer(GL.ELEMENT_ARRAY_BUFFER, this.elements);
    }
  };

  return Mesh;

})();

Material = (function() {
  Material.prototype.name = "?";

  Material.prototype.instances = null;

  function Material(name, shader) {
    this.name = name;
    this.instances = new Array(0);
    gfx.addMaterial(this);
    shader.addMaterial(this);
    return;
  }

  Material.prototype.register = function(primitive) {
    primitive.index = this.instances.length;
    this.instances[primitive.index] = primitive;
  };

  Material.prototype.remove = function(primitive) {
    var last, top;
    if (primitive.index === -1) {
      return;
    }
    last = this.instances.length - 1;
    if (primitive.index === last) {
      this.instances.pop();
    } else {
      top = this.instances.pop();
      top.index = primitive.index;
      this.instances[primitive.index] = top;
    }
    primitive.index = -1;
  };

  Material.prototype.clear = function() {
    var j, len, primitive, ref;
    ref = this.instances;
    for (j = 0, len = ref.length; j < len; j++) {
      primitive = ref[j];
      primitive.index = -1;
    }
    return this.instances = new Array(0);
  };

  return Material;

})();

Primitive = (function() {
  Primitive.prototype.pos = null;

  Primitive.prototype.height = 0;

  Primitive.prototype.scale = 1;

  Primitive.prototype.mesh = null;

  Primitive.prototype.tex = null;

  Primitive.prototype.color = vec4.fromValues(0, 0, 0, 1);

  Primitive.prototype.material = null;

  Primitive.prototype.index = -1;

  function Primitive(pos, mesh, material1, height, scale) {
    this.pos = pos;
    this.mesh = mesh;
    this.material = material1;
    if ((height != null)) {
      this.height = height;
    }
    if ((scale != null)) {
      this.scale = scale;
    }
    this.link();
    return;
  }

  Primitive.prototype.modelMatrix = function() {
    var model;
    model = mat4.create();
    return model = mat4.translate(model, model, vec3.fromValues(this.pos[0], this.pos[1], this.height));
  };

  Primitive.prototype.link = function() {
    if (this.index < 0) {
      this.material.register(this);
    }
  };

  Primitive.prototype.unlink = function() {
    this.material.remove(this);
  };

  return Primitive;

})();

OrthoCam = (function() {
  OrthoCam.prototype.pos = vec3.fromValues(0, 0, 1);

  OrthoCam.prototype.look = vec3.fromValues(0, 0, -1);

  OrthoCam.prototype.up = vec3.fromValues(0, 1, 0);

  OrthoCam.prototype.near = 0;

  OrthoCam.prototype.far = 10000;

  OrthoCam.prototype.zoom = 1;

  function OrthoCam(cPos, cLook, cUp) {
    if ((cPos != null)) {
      this.pos = cPos;
    } else {
      this.pos = vec3.clone(this.pos);
    }
    if ((cLook != null)) {
      this.look = cLook;
    } else {
      this.look = vec3.clone(this.look);
    }
    if (cUp) {
      this.up = cUp;
    } else {
      this.up = vec3.clone(this.up);
    }
    return;
  }

  OrthoCam.prototype.viewMatrix = function() {
    var antiPos, translate, view, xLocal, yLocal, zLocal;
    antiPos = vec3.create();
    vec3.negate(antiPos, this.pos);
    translate = mat4.create();
    mat4.fromTranslation(translate, antiPos);
    xLocal = vec3.create();
    yLocal = vec3.create();
    zLocal = vec3.create();
    view = mat4.create();
    vec3.cross(xLocal, this.look, this.up);
    vec3.normalize(xLocal, xLocal);
    vec3.cross(yLocal, xLocal, this.look);
    vec3.normalize(yLocal, yLocal);
    vec3.negate(zLocal, this.look);
    view[0] = xLocal[0];
    view[1] = yLocal[0];
    view[2] = zLocal[0];
    view[4] = xLocal[1];
    view[5] = yLocal[1];
    view[6] = zLocal[1];
    view[8] = xLocal[2];
    view[9] = yLocal[2];
    view[10] = zLocal[2];
    mat4.multiply(view, view, translate);
    return view;
  };

  OrthoCam.prototype.projMatrix = function(width, height) {
    var proj;
    width /= 2 * this.zoom;
    height /= 2 * this.zoom;
    proj = mat4.create();
    mat4.ortho(proj, -width, width, -height, height, this.far, this.near);
    return proj;
  };

  OrthoCam.prototype.viewRange = function(near, far) {
    this.near = near;
    this.far = far;
  };

  OrthoCam.prototype.lookAt = function(vx, y, z) {
    if (typeof vx === "number") {
      vx = vec3.fromValues(vx, y, z);
    }
    vec3.subtract(this.look, vx, this.pos);
    return vec3.normalize(this.look, this.look);
  };

  OrthoCam.prototype.lookIn = function(vx, y, z) {
    if (typeof vx === "number") {
      vx = vec3.fromValues(vx, y, z);
    }
    return vec3.normalize(this.look, vx);
  };

  OrthoCam.prototype.moveTo = function(vx, y, z) {
    if (typeof vx === "number") {
      vx = vec3.fromValues(vx, y, z);
    }
    return this.pos = vx;
  };

  OrthoCam.prototype.moveBy = function(vx, y, z) {
    if (typeof vx === "number") {
      vx = vec3.fromValues(vx, y, z);
    }
    return vec3.add(this.pos, this.pos, vx);
  };

  OrthoCam.prototype.orientUp = function(vx, y, z) {
    if (typeof vx === "number") {
      vx = vec3.fromValues(vx, y, z);
    }
    return vec3.normalize(this.up, vx);
  };

  OrthoCam.prototype.orientGravity = function(vx, y, z) {
    if (typeof vx === "number") {
      vx = vec3.fromValues(vx, y, z);
    }
    vec3.subtract(this.up, this.pos, vx);
    return vec3.normalize(this.up, this.up);
  };

  return OrthoCam;

})();

Camera = (function(superClass) {
  extend(Camera, superClass);

  function Camera() {
    return Camera.__super__.constructor.apply(this, arguments);
  }

  Camera.prototype.lens = glMatrix.toRadian(90);

  Camera.prototype.fov = Camera.lens;

  Camera.prototype.projMatrix = function(width, height) {
    var proj, ratio;
    ratio = width / height;
    proj = mat4.create();
    return mat4.perspective(proj, this.fov, ratio, this.near, this.far);
  };

  Camera.prototype.setZoom = function(zoom) {
    this.zoom = zoom;
    return this.fov = Math.atan(Math.tan(this.lens) / this.zoom);
  };

  Camera.prototype.setLens = function(lens) {
    this.lens = lens;
    this.fov = this.lens;
  };

  return Camera;

})(OrthoCam);
