#ApplyComponent: {
	#provider: "oam"
	#do:       "component-apply"
	cluster:   *"" | string
	value: {...}
	patch?: {...}
	...
}

#RenderComponent: {
	#provider: "oam"
	#do:       "component-render"
	cluster:   *"" | string
	value: {...}
	patch?: {...}
	output?: {...}
	outputs?: {...}
	...
}

#LoadComponets: {
	#provider: "oam"
	#do:       "load"
	...
}

#LoadPolicies: {
	#provider: "oam"
	#do:       "load-policies"
	value?: {...}
	...
}

#LoadComponetsInOrder: {
	#provider: "oam"
	#do:       "load-comps-in-order"
	...
}
