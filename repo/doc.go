package repo

/*
Src
	provides Templates
	generated Controls from teplates

	CACHE
	spakgs

Avail for build := Controls
Avail for install := Controls
Install no build := spakgs in cache


Src + Bin
	provides Templates
	provides PkgSet
	generated Controls from templates

	SEPARATE
	provides spakgs

	CACHE
	spakgs

Avail for build := Controls
Avail for install := Controls
Install no build := spakgs in cache and PkgSets

Bin
	provides PkgSet

	SEPARATE
	provides spakgs

Avail for build := none
Avail for install := Controls
Install no build := PkgSets

*/
