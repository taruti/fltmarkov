package main

//go:generate cpp -P -DNAME=OptimizeLinear256 -DSIZE=256 optimize_linear_src.go -o optimize_linear_256.go
//go:generate cpp -P -DNAME=OptimizeLinear64  -DSIZE=64  optimize_linear_src.go -o optimize_linear_64.go
//go:generate cpp -P -DNAME=OptimizeLinear16  -DSIZE=16  optimize_linear_src.go -o optimize_linear_16.go
