package graph

//type Crafting struct {
//	Crafting *game.Crafting
//	//Skill    Node
//	Items []Node
//}
//
//func NewCrafting(crafting *game.Crafting) Node {
//	var items []Node
//	for craftingItem, craftingQuantity := range crafting.Items {
//		items = append(items, NewItem(craftingItem, craftingQuantity))
//	}
//
//	return &Crafting{
//		Crafting: crafting,
//		//Skill:    NewSkill(crafting.Skill, crafting.Level),
//		Items: items,
//	}
//}
//
//func (c *Crafting) Type() Type {
//	return TypeCrafting
//}
//
//func (c *Crafting) Children() []Node {
//	//return append([]Node{c.Skill}, c.Items...)
//	return c.Items
//}
//
//func (c *Crafting) String() string {
//	return fmt.Sprintf("Craft %d item(s) into %d item(s)", len(c.Items), c.Crafting.Quantity)
//}
//
//func (c *Crafting) MarshalJSON() ([]byte, error) {
//	return marshal(c)
//}
