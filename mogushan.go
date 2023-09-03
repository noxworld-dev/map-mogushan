package mogushan

func OnEvent(typ string) {
	switch typ {
	case "MapInitialize":
		stoneGuard.Reset()
	}
}

func OnFrame() {
	stoneGuard.Update()
}
