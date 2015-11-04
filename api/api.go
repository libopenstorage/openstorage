/*
Package api defines the payload structures used between the client and the apiserver for the volumes and cluster packages.
*/
package api

type DriverType int

const (
	File = 1 << iota
	Block
	Object
	Clustered
	Graph
)
