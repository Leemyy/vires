// Generated by CoffeeScript 1.9.3
var Cell, Movement, Player, Random, convertGameCoords, convertMouseCoords, settings, vires,
  bind = function(fn, me){ return function(){ return fn.apply(me, arguments); }; };

settings = {
  minZoom: 1,
  maxZoom: 100,
  zoomSpeed: 0.2,
  factorLight: 0.3,
  factorDark: 0.4,
  indexNeutral: 10,
  indexSelf: 30,
  indexOther: 20,
  indexMarker: 50,
  offsetMovement: -100,
  gauge: 0.1
};

vires = {
  room: 0,
  time: 0,
  timePrev: 0,
  timeDelta: 0,
  Self: 0,
  states: {
    loading: {},
    lobby: {},
    match: {},
    noConnection: {},
    debug: {}
  },
  active: null,
  next: {
    state: null,
    data: null
  },
  init: function() {
    this.active = this.states.loading;
    this.active.load();
  },
  load: function(stateName, data) {
    this.next.state = this.states[stateName];
    return this.next.data = data;
  }
};

vires.states.match = {
  players: null,
  cells: null,
  movements: null,
  lookup: null,
  selection: null,
  targetMarker: null,
  target: null,
  random: null,
  timeStart: 0,
  fieldSize: vec2.fromValues(800, 800),
  spectating: false,
  maxCellSize: 1,
  minCellSize: 10000,
  cameraDrag: false,
  cameraStart: null,
  load: function(Field) {
    var cellData, firstCell, i, l, len, m, owner, palette, radius, ref, ref1, server, start;
    server = new Player(0);
    server.color = gfx.makeColor(0);
    this.players = {
      0: server
    };
    this.cells = new Array(Field.Cells.length);
    this.lookup = new Array(Field.Cells.length);
    this.movements = {};
    this.selection = {};
    this.timeStart = vires.time;
    this.targetMarker = new Primitive(vec2.create(), gfx.mesh.target, gfx.material.marker, settings.indexMarker);
    this.targetMarker.unlink();
    this.fieldSize = vec2.fromValues(Field.Size.X, Field.Size.Y);
    ref = Field.Cells;
    for (l = 0, len = ref.length; l < len; l++) {
      cellData = ref[l];
      this.cells[cellData.ID] = new Cell(cellData);
      radius = cellData.Body.Radius;
      if (radius > this.maxCellSize) {
        this.maxCellSize = radius;
      }
      if (radius < this.minCellSize) {
        this.minCellSize = radius;
      }
    }
    this.lookup = this.cells.slice(0);
    this.lookup.sort(function(a, b) {
      return a.Pos[x] - b.Pos[x];
    });
    this.random = new Random(vires.room);
    this.random.next();
    this.random.seed *= Math.floor(this.cells[0].Radius);
    this.random.next();
    this.random.seed += Math.floor(Math.pow(this.cells[0].Pos[0], this.cells[0].Pos[1]));
    this.random.next();
    palette = gfx.color.slice(1);
    this.random.shuffle(palette);
    this.spectating = true;
    firstCell = null;
    for (i = m = 0, ref1 = Field.StartCells.length; 0 <= ref1 ? m < ref1 : m > ref1; i = 0 <= ref1 ? ++m : --m) {
      start = Field.StartCells[i];
      owner = new Player(start.Owner);
      owner.color = vec4.clone(palette[i % palette.length]);
      this.players[start.Owner] = owner;
      this.cells[start.Cell].switchOwner(owner);
      if (owner.ID === vires.Self) {
        this.spectating = false;
        firstCell = this.cells[start.Cell];
      }
    }
    this.clearMarkers();
    settings.minZoom = 800 / this.fieldSize[1];
    settings.maxZoom = 200 / this.minCellSize;
    if (this.spectating) {
      vec2.set(gfx.camera.pos, this.fieldSize[x] / 2, this.fieldSize[y] / 2);
      gfx.camera.zoom = gfx.width / this.fieldSize[x];
    } else {
      vec2.set(gfx.camera.pos, firstCell.Pos[x], firstCell.Pos[y]);
      gfx.camera.zoom = 100 / firstCell.Radius;
    }
  },
  unload: function() {
    gfx.material.cell.clear();
    gfx.material.movement.clear();
    gfx.material.marker.clear();
  },
  digestInput: function() {
    var delta, hover, id, lerp, marked, offset, prevZoom, ref, sources, zoomFactor;
    if (input.left && !this.spectating) {
      hover = this.cellAt(input.cursor);
      if ((hover != null)) {
        if (hover.Owner === vires.Self) {
          this.selection[hover.ID] = hover;
        }
        if (!(this.target != null)) {
          this.target = hover;
          this.targetMarker.pos = hover.Pos;
          this.targetMarker.scale = hover.Radius;
          this.targetMarker.link();
          hover.unmark();
        } else if (this.target.ID !== hover.ID) {
          if (this.target.Owner === vires.Self) {
            this.target.mark();
          }
          this.target = hover;
          this.targetMarker.pos = hover.Pos;
          this.targetMarker.scale = hover.Radius;
          hover.unmark();
        }
      } else if ((this.target != null)) {
        if (this.target.Owner === vires.Self) {
          this.target.mark();
        }
        this.target = null;
        this.targetMarker.unlink();
      }
    }
    if (input.leftReleased && !this.spectating) {
      hover = this.cellAt(input.cursor);
      if ((hover != null)) {
        sources = [];
        ref = this.selection;
        for (id in ref) {
          marked = ref[id];
          if (marked.ID !== hover.ID) {
            sources.push(marked);
          }
        }
        connection.sendMove(hover, sources);
      }
      this.target = null;
      this.selection = {};
      gfx.material.marker.clear();
    }
    if (input.rightPressed) {
      this.cameraDrag = true;
      this.cameraStart = input.cursor;
    }
    if (input.rightReleased) {
      this.cameraDrag = false;
    }
    if (this.cameraDrag) {
      delta = vec2.subtract(vec2.create(), this.cameraStart, input.cursor);
      vec2.add(gfx.camera.pos, gfx.camera.pos, delta);
    }
    if (input.scroll !== 0 && !this.cameraDrag) {
      prevZoom = gfx.camera.zoom;
      if (input.scroll > 0) {
        gfx.camera.zoom *= 1 + input.scroll * settings.zoomSpeed;
        if (gfx.camera.zoom > settings.maxZoom) {
          gfx.camera.zoom = settings.maxZoom;
        }
      } else {
        gfx.camera.zoom /= 1 - input.scroll * settings.zoomSpeed;
        if (gfx.camera.zoom < settings.minZoom) {
          gfx.camera.zoom = settings.minZoom;
        }
      }
      zoomFactor = gfx.camera.zoom / prevZoom;
      offset = vec2.subtract(vec2.create(), input.cursor, gfx.camera.pos);
      lerp = 1 - 1 / zoomFactor;
      vec2.scale(offset, offset, lerp);
      vec2.add(gfx.camera.pos, gfx.camera.pos, offset);
    }
    vec2.max(gfx.camera.pos, gfx.camera.pos, vec2.create());
    vec2.min(gfx.camera.pos, gfx.camera.pos, this.fieldSize);
  },
  digestTraffic: function() {
    var A, B, Msg, cell, data, id, l, len, ref, update;
    Msg = connection.messages.pop();
    while ((Msg != null)) {
      data = Msg.Data;
      switch (Msg.Type) {
        case "Movement":
          this.movements[data.ID] = new Movement(data);
          break;
        case "Replication":
          for (l = 0, len = data.length; l < len; l++) {
            update = data[l];
            this.cells[update.ID].update(update.Stationed);
          }
          break;
        case "Conflict":
          this.movements[data.Movement].kill();
          delete this.movements[data.Movement];
          this.cells[data.Cell.ID].update(data.Cell.Stationed);
          this.cells[data.Cell.ID].switchOwner(this.players[data.Cell.Owner]);
          break;
        case "Collision":
          A = this.movements[data.A.ID];
          B = this.movements[data.B.ID];
          if (data.A.Moving > 0) {
            A.update(data.A);
          } else {
            A.kill();
            delete this.movements[A.ID];
          }
          if (data.B.Moving > 0) {
            B.update(data.B);
          } else {
            B.kill();
            delete this.movements[B.ID];
          }
          break;
        case "EliminatedPlayer":
          if (vires.Self === data) {
            this.spectating = true;
            ref = this.selection;
            for (id in ref) {
              cell = ref[id];
              cell.unmark();
            }
            this.targetMarker.unlink();
          }
          this.killPlayer(data);
          break;
        case "Winner":
          vires.load("lobby", this.players[data].color);
          break;
        default:
          connection.defaultDigest(Msg);
      }
      Msg = connection.messages.pop();
    }
  },
  animate: function() {
    var k, mov, ref, time;
    time = vires.time;
    ref = this.movements;
    for (k in ref) {
      mov = ref[k];
      mov.move(time);
    }
  },
  clearMarkers: function() {
    gfx.material.marker.clear();
  },
  killPlayer: function(ID) {
    var cell, l, len, len1, m, move, ref, ref1;
    this.players[ID].kill();
    ref = this.cells;
    for (l = 0, len = ref.length; l < len; l++) {
      cell = ref[l];
      if (cell.Owner === ID) {
        cell.Owner = 0;
      }
    }
    ref1 = this.movements;
    for (m = 0, len1 = ref1.length; m < len1; m++) {
      move = ref1[m];
      if (move.Owner === ID) {
        move.kill();
        delete this.movements[move.ID];
      }
    }
  },
  cellAt: function(pos) {
    var cell, l, len, ref;
    ref = this.lookup;
    for (l = 0, len = ref.length; l < len; l++) {
      cell = ref[l];
      if (vec2.distance(input.cursor, cell.Pos) <= cell.Radius) {
        return cell;
      }
    }
    return null;
  },
  cellAtFast: function(pos) {
    var current, head, highBound, i, l, lowBound, max, min, ref, ref1, searching, targetMax, targetMin;
    targetMin = pos[x] - this.maxCellSize;
    targetMax = pos[x] + this.maxCellSize;
    lowBound = this.lookup.length - 1;
    highBound = 0;
    min = 0;
    max = this.lookup.length - 1;
    searching = true;
    while (max - min >= 0) {
      head = min + Math.ceil((max - min) / 2);
      current = this.lookup[head].Pos[x];
      if (current - targetMin < 0) {
        min = head + 1;
        if (highBound < head) {
          highBound = head;
        }
      } else {
        max = head - 1;
        if (targetMax - current < 0) {
          if (lowBound > head) {
            lowBound = head;
          }
        } else {
          if (highBound < head) {
            highBound = head;
          }
          if (lowBound < head) {
            lowBound = head;
          }
        }
      }
    }
    min = Math.max(highBound, lowBound) + 1;
    max = this.lookup.length - 1;
    while (max - min >= 0) {
      head = min + Math.ceil((max - min) / 2);
      current = this.lookup[head].Pos[x];
      if (targetMax - current < 0) {
        max = head - 1;
        if (highBound < head) {
          highBound = head;
        }
      } else {
        min = head + 1;
        if (!(current - targetMin < 0)) {
          if (highBound < head) {
            highBound = head;
          }
        }
      }
    }
    if (highBound - lowBound >= 0) {
      for (i = l = ref = lowBound, ref1 = highBound; ref <= ref1 ? l <= ref1 : l >= ref1; i = ref <= ref1 ? ++l : --l) {
        if (vec2.distance(input.cursor, this.lookup[i].Pos) <= this.lookup[i].Radius) {
          return this.lookup[i];
        }
      }
    }
    return null;
  }
};

vires.states.loading = {
  animation: [],
  load: function() {
    var color, i, l, material, mesh, segment;
    vec2.set(gfx.camera.pos, 0, 0);
    gfx.camera.zoom = 5;
    color = gfx.makeColor(2);
    mesh = gfx.mesh.round;
    material = gfx.material.loading;
    this.animation = new Array(10);
    for (i = l = 0; l < 10; i = ++l) {
      this.animation[i] = vec2.create();
      segment = new Primitive(this.animation[i], mesh, material, i);
      segment.scale = 0.5 + i * 0.05;
      segment.color = vec4.clone(color);
      segment.color[3] = 0.1 + i * 0.1;
    }
  },
  unload: function() {
    gfx.material.loading.clear();
  },
  digestInput: function() {},
  digestTraffic: function() {},
  animate: function() {
    var distance, i, l, step;
    for (i = l = 0; l < 9; i = ++l) {
      vec2.copy(this.animation[i], this.animation[i + 1]);
    }
    distance = vec2.distance(input.cursor, this.animation[9]);
    if (distance !== 0) {
      step = Math.min(2, distance);
      step = step / distance;
      vec2.lerp(this.animation[9], this.animation[9], input.cursor, step);
    }
  }
};

vires.states.lobby = {
  animation: [],
  load: function(winner) {
    var color, i, l, material, mesh, segment;
    vec2.set(gfx.camera.pos, 0, 0);
    gfx.camera.zoom = 5;
    color = null;
    if ((winner != null)) {
      color = winner;
    } else {
      color = gfx.makeColor(10);
    }
    mesh = gfx.mesh.round;
    material = gfx.material.loading;
    this.animation = new Array(10);
    for (i = l = 0; l < 10; i = ++l) {
      this.animation[i] = vec2.create();
      segment = new Primitive(this.animation[i], mesh, material, i);
      segment.color = vec4.clone(color);
      segment.color[3] = 0.1 + i * 0.1;
    }
  },
  unload: function() {
    gfx.material.loading.clear();
  },
  digestInput: function() {},
  digestTraffic: function() {
    var Msg;
    Msg = connection.messages.pop();
    while ((Msg != null)) {
      connection.defaultDigest(Msg);
      Msg = connection.messages.pop();
    }
  },
  animate: vires.states.loading.animate
};

vires.states.noConnection = {
  load: function() {},
  unload: function() {},
  digestInput: function() {},
  digestTraffic: function() {},
  animate: function() {}
};

vires.states.debug = {
  load: function() {},
  unload: function() {},
  digestInput: function() {},
  digestTraffic: function() {},
  animate: function() {}
};

convertGameCoords = function(pos) {
  var out;
  out = vec2.subtract(vec2.create(), pos, gfx.camera.pos);
  vec2.scale(out, out, gfx.camera.zoom);
  out[1] = -out[1];
  out[0] += gfx.width / 2;
  out[1] += gfx.height / 2;
  return out;
};

convertMouseCoords = function(pos) {
  var out;
  out = vec2.fromValues(pos[0] - gfx.width / 2, gfx.height / 2 - pos[1]);
  vec2.scale(out, out, 1 / gfx.camera.zoom);
  vec2.add(out, out, gfx.camera.pos);
  return out;
};

Random = (function() {
  Random.prototype.seed = 0;

  function Random(seed) {
    this.seed = seed;
    this.nextIn = bind(this.nextIn, this);
    this.next = bind(this.next, this);
    return;
  }

  Random.prototype.next = function() {
    this.seed = (this.seed * 9301 + 49297) % 233280;
    return this.seed / 233280;
  };

  Random.prototype.nextIn = function(min, max) {
    var rnd;
    this.seed = (this.seed * 9301 + 49297) % 233280;
    rnd = this.seed / 233280;
    return min + rnd * (max - min);
  };

  Random.prototype.shuffle = function(arr) {
    var i, j, t;
    i = arr.length;
    while (--i > 0) {
      j = ~~(this.next() * (i + 1));
      t = arr[j];
      arr[j] = arr[i];
      arr[i] = t;
    }
    return arr;
  };

  return Random;

})();

Player = (function() {
  Player.prototype.ID = 0;

  Player.prototype.color = null;

  Player.prototype.colorLight = null;

  Player.prototype.colorDark = null;

  Player.prototype.alive = true;

  function Player(ID1) {
    this.ID = ID1;
    this.alive = true;
    this.color = gfx.color[0];
    this.colorDark = vec4.lerp(vec4.create(), this.color, gfx.black, settings.factorDark);
    this.colorLight = vec4.lerp(vec4.create(), this.color, gfx.white, settings.factorLight);
    return;
  }

  Player.prototype.swapColor = function(color) {
    vec4.copy(this.color, color);
    vec4.lerp(this.colorLight, this.color, gfx.white, settings.factorLight);
    return vec4.lerp(this.colorDark, this.color, gfx.black, settings.factorDark);
  };

  Player.prototype.kill = function() {
    vec4.copy(this.color, gfx.color[0]);
    return this.alive = false;
  };

  return Player;

})();

Cell = (function() {
  Cell.prototype.ID = 0;

  Cell.prototype.Owner = 0;

  Cell.prototype.Pos = vec2.fromValues(0, 0);

  Cell.prototype.Radius = 1;

  Cell.prototype.Capacity = 1;

  Cell.prototype.Stationed = 0;

  Cell.prototype.body = null;

  Cell.prototype.gauge = null;

  Cell.prototype.antigauge = null;

  Cell.prototype.marker = null;

  function Cell(Data) {
    this.ID = Data.ID;
    this.Pos = vec2.fromValues(Data.Body.Location.X, Data.Body.Location.Y);
    this.Radius = Data.Body.Radius;
    this.Capacity = Data.Capacity;
    this.body = new Primitive(this.Pos, gfx.mesh.round, gfx.material.cell);
    this.gauge = new Primitive(this.Pos, gfx.mesh.round, gfx.material.cell);
    this.antigauge = new Primitive(this.Pos, gfx.mesh.round, gfx.material.cell);
    this.marker = new Primitive(this.Pos, gfx.mesh.mark, gfx.material.marker);
    this.switchHeight(settings.indexNeutral);
    this.body.color = gfx.color[0];
    this.gauge.color = gfx.black;
    this.antigauge.color = gfx.color[0];
    this.body.scale = this.Radius;
    this.marker.scale = this.Radius;
    this.gauge.scale = 0;
    this.antigauge.scale = 0;
    return;
  }

  Cell.prototype.switchOwner = function(owner) {
    this.Owner = owner.ID;
    if (this.Owner === vires.Self) {
      this.switchHeight(settings.indexSelf);
    } else if (this.Owner === 0) {
      this.switchHeight(settings.indexNeutral);
    } else {
      this.switchHeight(settings.indexOther);
    }
    this.body.color = owner.color;
    this.antigauge.color = owner.color;
  };

  Cell.prototype.switchHeight = function(height) {
    this.body.height = height;
    this.gauge.height = height + 1;
    this.antigauge.height = height + 2;
    return this.marker.height = height + 5;
  };

  Cell.prototype.update = function(stationed) {
    var fullnes, trailing;
    this.Stationed = stationed;
    fullnes = stationed / this.Capacity;
    fullnes = Math.sqrt(fullnes);
    trailing = Math.max(fullnes - settings.gauge, 0);
    fullnes *= this.Radius;
    trailing *= this.Radius;
    this.gauge.scale = fullnes;
    this.antigauge.scale = trailing;
  };

  Cell.prototype.mark = function() {
    this.marker.link();
  };

  Cell.prototype.unmark = function() {
    this.marker.unlink();
  };

  return Cell;

})();

Movement = (function() {
  Movement.prototype.ID = 0;

  Movement.prototype.Owner = 0;

  Movement.prototype.Moving = 0;

  Movement.prototype.O = vec2.fromValues(0, 0);

  Movement.prototype.Radius = 1;

  Movement.prototype.V = vec2.fromValues(0, 0);

  Movement.prototype.birth = 0;

  Movement.prototype.pos = vec2.fromValues(0, 0);

  Movement.prototype.body = null;

  function Movement(Data) {
    this.ID = Data.ID;
    this.Owner = Data.Owner;
    this.Moving = Data.Moving;
    this.O = vec2.fromValues(Data.Body.Location.X, Data.Body.Location.Y);
    this.Radius = Data.Body.Radius;
    this.V = vec2.fromValues(Data.Direction.X, Data.Direction.Y);
    this.birth = vires.time;
    this.pos = vec2.clone(this.O);
    this.body = new Primitive(this.pos, gfx.mesh.round, gfx.material.movement);
    if (this.Owner === vires.Self) {
      this.body.height = settings.indexSelf + settings.offsetMovement;
    } else if (this.Owner === 0) {
      this.body.height = settings.indexNeutral + settings.offsetMovement;
    } else {
      this.body.height = settings.indexOther + settings.offsetMovement;
    }
    this.body.scale = this.Radius;
    this.body.color = vires.states.match.players[this.Owner].color;
    return;
  }

  Movement.prototype.move = function(now) {
    vec2.scaleAndAdd(this.pos, this.O, this.V, now - this.birth);
  };

  Movement.prototype.kill = function() {
    this.body.unlink();
  };

  Movement.prototype.update = function(Data) {
    this.Moving = Data.Moving;
    vec2.set(this.O, Data.Body.Location.X, Data.Body.Location.Y);
    this.Radius = Data.Body.Radius;
    vec2.set(this.V, Data.Direction.X, Data.Direction.Y);
    this.birth = vires.time;
    vec2.copy(this.pos, this.O);
    this.body.scale = this.Radius;
  };

  return Movement;

})();
