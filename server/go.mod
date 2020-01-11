module github.com/simonfrey/unshort.link

go 1.13

require (
	github.com/pkg/errors v0.8.1
	github.com/sergi/go-diff v1.1.0
	github.com/sirupsen/logrus v1.4.2
	unshort.link/db v0.0.0
)

replace unshort.link/db v0.0.0 => ./db
