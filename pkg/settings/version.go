package settings

var (
	version = "dev"
)

func IsDevelop() bool {
	return "dev" == version
}
