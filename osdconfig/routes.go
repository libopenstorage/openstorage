package osdconfig

import "path/filepath"

// GetRoutes provides HTTP endpoints to be used with HTTP routers
func (manager *configManager) GetRoutes() []*Routes {
	R := make([]*Routes, 0, 0)

	r := new(Routes)
	r.Method = Get
	r.Path = filepath.Join(BasePath, List)
	r.Fn, _ = manager.httpFunc(r.Method, manager.GetNodeList)
	R = append(R, r)

	r = new(Routes)
	r.Method = Get
	r.Path = filepath.Join(BasePath, Cluster)
	r.Fn, _ = manager.httpFunc(r.Method, manager.GetClusterConf)
	R = append(R, r)

	r = new(Routes)
	r.Method = Get
	r.Path = filepath.Join(BasePath, Node, string2tag(Id))
	r.Fn, _ = manager.httpFunc(r.Method, manager.GetNodeConf)
	R = append(R, r)

	r = new(Routes)
	r.Method = Put
	r.Path = filepath.Join(BasePath, Cluster)
	r.Fn, _ = manager.httpFunc(r.Method, manager.SetClusterConf)
	R = append(R, r)

	r = new(Routes)
	r.Method = Put
	r.Path = filepath.Join(BasePath, Node)
	r.Fn, _ = manager.httpFunc(r.Method, manager.SetNodeConf)
	R = append(R, r)

	r = new(Routes)
	r.Method = Post
	r.Path = filepath.Join(BasePath, Cluster)
	r.Fn, _ = manager.httpFunc(r.Method, manager.SetClusterConf)
	R = append(R, r)

	r = new(Routes)
	r.Method = Post
	r.Path = filepath.Join(BasePath, Node)
	r.Fn, _ = manager.httpFunc(r.Method, manager.SetNodeConf)
	R = append(R, r)

	r = new(Routes)
	r.Method = Del
	r.Path = filepath.Join(BasePath, Node, string2tag(Id))
	r.Fn, _ = manager.httpFunc(r.Method, manager.DeleteNodeConf)
	R = append(R, r)

	return R
}

func string2tag(name string) string {
	return "{" + name + "}"
}
