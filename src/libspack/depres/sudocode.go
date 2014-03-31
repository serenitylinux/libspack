func DepCheck(c *pkgdep.PkgDep, base *pkgdep.PkgDep, globalflags *flagconfig.FlagList, forge_deps *pkgdep.PkgDepList, wield_deps *pkgdep.PkgDepList, missing *MissingInfoList, params DepResParams) bool {
	/*
	I am a node
	I have parents
	if I am satisfied(){ // I am satisfied iff I am installed correctly or I will be installed correctly
		YAY!
	} else {
		// I need to change/be installed
		case 1: I will be installed but not correctly
		case 2: I am installed but not correctly
		case 3: I need to be installed
		
		What do I need to be installed correctly?
		I need to check my children now to make sure they are installed correctly...
		
	} 
	
	*/
	
	// If the current node is somewhere in the path (or all?), go back to that spot in order to re-evaluate to make sure everything is honkydoory
	
	/*
	path stack
	if Installed or will be installed (with correct flags)
	{
		continue	
	}
	else
	{
		add to install list
	}
	tobebuilt[]
	for node in the install list
	{
		deps = get the depslist of node
		for dep in deps
			if dep is installed
				yay!
			else if dep is built
				basically as good
				push onto path to check the deps
			else //needs to be built
				//time to setup the subtrees for the deps that need to be built
				if the dep is in the path
					crash out
				else
					tobebuilt.add(dep)
	
	}
	for node in tobebuilt
	{
		depcheck(node .........)
	}
	
	struct node {
		dirty
		control
		repo
		flags
		[] *children
		[] *parents
		
		*mybuildtree
	}
	
	
	
	func BuildTree(self, tree, isbuild) {
		
		if not dirty && Installed or will be installed (with correct flags)
		{
			return pointer to that
		}
		
		mark self not dirty
		
		for dep in deps or bdeps depending on isbuild {
			depnode := find dep in current tree
			if depnode is in current tree {
				modify dep to fit our criteria
				if we can't do that {
					die
				}
				mark dep as dirty!!!!
			} else {
				create dep node
			}
			
			update references from self to depnode and vice versa
			
			BuildTree(depnode, base, false)
		}
	}
	
	func FindToBuild(tree, treelist) {
		list of packages to build
		
		for each node in tree {
			if node.NoSpakgExists {
				list.add(node)
			}
		}
		
		for node in list {
			if !treelist.contains(node) {
				node.MyBuildTree = node.Copy()
				BuildTree(node.MyBuildTree, node.MyBuildTree, true)
				treelist.prepend(node.MyBuildTree)
				FindToBuild(node.MyBuildTree, treelist)
			} else {
				//TODO
			}
		}
	}
	
	
	
	*/

}