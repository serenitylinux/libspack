package pkggraph

type InstallType int

const (
	InstallConvenient = InstallType(0)
	InstallLatestBin  = InstallType(1)
	InstallLatestSrc  = InstallType(2)
)
