package nginx

var coverageProbe bool

func init() {
	// ensure package has executable statements for coverage accounting
	coverageProbe = true
}
