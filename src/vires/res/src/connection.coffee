connection=
	version: "0.1"
	url: "ws://" + window.location.host + "/#{vires.room}/c"
	messages: new Array(0)
	socket: null

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
				connection.messages.push(JSON.parse(msg.data))
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


packets=
	field:{
    "Type": "Field",
    "Version": "0.1",
    "Data": {
        "Cells": [
                {
                        "ID": 0,
                        "Body": {
                                "Location": {
                                        "X": 200,
                                        "Y": 200
                                },
                                "Radius": 3
                        },
                        "Capacity": 10
                },
                {
                        "ID": 1,
                        "Body": {
                                "Location": {
                                        "X": 180,
                                        "Y": 210
                                },
                                "Radius": 5
                        },
                        "Capacity": 10
                },
                {
                        "ID": 2,
                        "Body": {
                                "Location": {
                                        "X": 200,
                                        "Y": 260
                                },
                                "Radius": 3
                        },
                        "Capacity": 10
                },
                {
                        "ID": 3,
                        "Body": {
                                "Location": {
                                        "X": 240,
                                        "Y": 190
                                },
                                "Radius": 5
                        },
                        "Capacity": 10
                },
                {
                        "ID": 4,
                        "Body": {
                                "Location": {
                                        "X": 230,
                                        "Y": 230
                                },
                                "Radius": 4
                        },
                        "Capacity": 10
                }
        ],
        "StartCells": [
                {
                        "Owner": 1,
                        "Cell": 3
                },
                {
                        "Owner": 2,
                        "Cell": 2
                }
        ],
        "Size": {
                "X": 800,
                "Y": 600
        }
    }
}