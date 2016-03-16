package cfg

import (
	"github.com/mhuisi/fcfg/jsoncfg"
	"github.com/mhuisi/logg"
)

type GeneralCfg struct {
	IP             string
	UseTLS         bool
	PrivateKeyPath string
	CertPath       string
	DebugLogging   bool
}

type MapgenCfg struct {
	Gap                float64
	MinRadius          float64
	MaxRadius          float64
	MinStartCellRadius float64
	CellsNearStartCell int
	MinCellsPerPlayer  int
	MaxCellsPerPlayer  int
}

type GameplayCfg struct {
	ReplicationInterval float64
	StartStationed      float64
	Capacity            float64
	Replication         float64
	NeutralReplication  float64
	MovementRadius      float64
	MovementSpeed       float64
}

type config struct {
	General  GeneralCfg
	Mapgen   MapgenCfg
	Gameplay GameplayCfg
}

func (c *config) Copy() interface{} {
	clone := *c
	return &clone
}

var (
	defaultCfg = &config{
		General: GeneralCfg{
			IP: "localhost",
		},
		Mapgen: MapgenCfg{
			Gap:                80,
			MinRadius:          80,
			MaxRadius:          160,
			MinStartCellRadius: 120,
			CellsNearStartCell: 4,
			MinCellsPerPlayer:  10,
			MaxCellsPerPlayer:  15,
		},
		Gameplay: GameplayCfg{
			ReplicationInterval: 1.5,
			StartStationed:      0.3,
			Capacity:            0.005,
			Replication:         0.2,
			NeutralReplication:  0.5,
			MovementRadius:      10,
			MovementSpeed:       3000,
		},
	}
	General  GeneralCfg
	Mapgen   MapgenCfg
	Gameplay GameplayCfg
)

func validationErr(s string, v ...interface{}) {
	logg.Fatal("Error while validating config: "+s, v...)
}

func init() {
	loaded, err := jsoncfg.New("cfg.json", defaultCfg).Load()
	if err != nil {
		logg.Fatal("Cannot load config file: %s", err)
	}
	c := loaded.(*config)
	General = c.General
	Mapgen = c.Mapgen
	if Mapgen.Gap < 0 {
		validationErr("Gap may not be negative.")
	}
	if Mapgen.MinRadius < 0 {
		validationErr("MinRadius may not be negative.")
	}
	if Mapgen.MaxRadius < Mapgen.MinRadius {
		validationErr("MaxRadius may not be smaller than MinRadius.")
	}
	if Mapgen.MinStartCellRadius > Mapgen.MaxRadius {
		validationErr("MinStartCellRadius may not be larger than MaxRadius.")
	}
	if Mapgen.CellsNearStartCell < 0 || Mapgen.CellsNearStartCell > 8 {
		validationErr("CellsNearStartCell must be a value between 0 and 8 (for every value larger than 8 you're running risk of never finishing the map generation).")
	}
	if Mapgen.MinCellsPerPlayer < 0 {
		validationErr("MinCellsPerPlayer may not be negative.")
	}
	if Mapgen.MaxCellsPerPlayer < Mapgen.MinCellsPerPlayer {
		validationErr("MaxCellsPerPlayer may not be larger than MinCellsPerPlayer.")
	}
	Gameplay = c.Gameplay
	if Gameplay.ReplicationInterval < 0.001 {
		validationErr("ReplicationInterval may not be smaller than 10ms.")
	}
	if Gameplay.StartStationed < 0 || Gameplay.StartStationed > 1 {
		validationErr("StartStationed must be a value between 0 and 1.")
	}
	if Gameplay.Capacity < 0 {
		validationErr("Capacity may not be negative.")
	}
	if Gameplay.Replication < 0 {
		validationErr("Replication may not be negative.")
	}
	if Gameplay.NeutralReplication < 0 {
		validationErr("NeutralReplication may not be negative.")
	}
	if Gameplay.MovementRadius < 0 {
		validationErr("MovementRadius may not be negative.")
	}
	if Gameplay.MovementSpeed < 0 {
		validationErr("Speed may not be negative.")
	}
	logg.Info("Configuration loaded.")
}
