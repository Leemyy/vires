connection=
	version: "0.1"
	url: "ws://" + window.location.host + "/#{vires.room}/c"
	messages: new Array(0)
	socket: null
	debug: new Array(0)

	init: ->
		connection.socket = new WebSocket(@url)

		connection.socket.onopen= (event)->
			vires.load("lobby")
			return

		connection.socket.onerror= (event)->
			#...
			return

		connection.socket.onmessage= (msg)->
			try
				Packet = JSON.parse(msg.data)
				connection.messages.unshift(Packet)
				#if (Packet.Type != "Replication")
				#	connection.debug.unshift(msg.data)
			catch err
				console.error(err)
				#...
			return

		connection.socket.onclose= (closed)->
			vires.load("noConnection")
			return
		
		return

	send: (type, payload)->
		packet = 
			Type: type
			Version: @version
			Data: payload
		data = JSON.stringify(packet)
		@socket.send(data)
		return


	sendMove: (target, sources)->
		for source in sources
			move =
				Source: source.ID
				Dest: target.ID
			@send("Movement", move)
		return


	defaultDigest: (Msg)->
		switch Msg.Type
			when "Movement" #
				return
			when "Replication" #
				return
			when "Conflict" #
				return
			when "Collision" #
				return
			when "EliminatedPlayer" #
				return
			when "Winner"
				return
			when "Field" #
				vires.load("match", Msg.Data)
			when "Join" #
				#Do nuffin
				return
			when "OwnID" #
				vires.Self = Msg.Data
		return
