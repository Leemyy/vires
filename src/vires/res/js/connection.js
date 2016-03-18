// Generated by CoffeeScript 1.9.3
var connection;

connection = {
  version: "0.1",
  url: "ws://" + window.location.host + ("/" + vires.room + "/c"),
  messages: new Array(0),
  socket: null,
  debug: new Array(0),
  init: function() {
    connection.socket = new WebSocket(this.url);
    connection.socket.onopen = function(event) {
      vires.load("lobby");
    };
    connection.socket.onerror = function(event) {};
    connection.socket.onmessage = function(msg) {
      var Packet, err;
      try {
        Packet = JSON.parse(msg.data);
        connection.messages.unshift(Packet);
      } catch (_error) {
        err = _error;
        console.error(err);
      }
    };
    connection.socket.onclose = function(closed) {
      vires.load("noConnection");
    };
  },
  send: function(type, payload) {
    var data, packet;
    packet = {
      Type: type,
      Version: this.version,
      Data: payload
    };
    data = JSON.stringify(packet);
    this.socket.send(data);
  },
  sendMove: function(target, sources) {
    var i, len, move, source;
    for (i = 0, len = sources.length; i < len; i++) {
      source = sources[i];
      move = {
        Source: source.ID,
        Dest: target.ID
      };
      this.send("Movement", move);
    }
  },
  defaultDigest: function(Msg) {
    switch (Msg.Type) {
      case "Movement":
        return;
      case "Replication":
        return;
      case "Conflict":
        return;
      case "Collision":
        return;
      case "EliminatedPlayer":
        return;
      case "Winner":
        return;
      case "Field":
        vires.load("match", Msg.Data);
        break;
      case "Join":
        return;
      case "OwnID":
        vires.Self = Msg.Data;
    }
  }
};
