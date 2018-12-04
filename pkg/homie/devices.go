package homie

type Device struct {
	Name  string
	Nodes []*Node
}

// func (e *Device) AddNode(node Node) *Node {
// 	e.Nodes = append(e.Nodes, node)
// 	return &node
// }

type Node struct {
	Name       string
	Properties []*Property
}

// func (e *Node) AddProperty(property Property) *Property {
// 	e.Properties = append(e.Properties, property)
// 	return &property
// }

type Property struct {
	Name string
}
