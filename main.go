package main

import (
	"context"
	"errors"
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

var ErrCraftingNoMoreMaterials = errors.New("no more crafting materials")
var ErrBankInsufficientQuantity = errors.New("bank has insufficient quantity")

type characterAction struct {
	Character *Character
	Action    string
	// TODO: Color?
}

type EndCondition interface {
	ShouldEnd() bool
}

type QuantityEndCondition int

func (e QuantityEndCondition) ShouldEnd(completedQuantity int) bool {
	return completedQuantity == int(e)
}

type CharacterTask struct {
	Type         string // harvest, craft, item, task, idle
	Code         string
	Quantity     int
	EndCondition EndCondition // TODO: use this in place of quantity so we can get more creative
}

func main() {
	if err := ui.Init(); err != nil {
		log.Fatal(err)
	}
	defer ui.Close()

	c, err := client.NewClientWithResponses("https://api.artifactsmmo.com", client.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+os.Getenv("ARTIFACTS_TOKEN"))
		return nil
	}))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	var characterNames []string
	for i := 1; i <= 5; i++ {
		name := fmt.Sprintf("curlyBoy%d", i)
		if i == 1 {
			name = "curlyBoy"
		}
		characterNames = append(characterNames, name)
	}

	levelNames := []string{"combat", "woodcutting", "mining", "fishing", "jewelrycrafting"}
	colors := []ui.Color{ui.ColorBlue, ui.ColorGreen, ui.ColorCyan, ui.ColorYellow, ui.ColorMagenta}

	tasks := []CharacterTask{
		{Type: "task_when_not_crafting", Code: "copper_ring"},
		{Type: "harvest", Code: "spruce_tree", Quantity: math.MaxUint32},
		{Type: "task_when_not_crafting", Code: "copper"},
		{Type: "harvest", Code: "copper_rocks", Quantity: math.MaxUint32},
		{Type: "task_when_not_crafting", Code: "spruce_plank"},

		//{Type: "fight", Code: "blue_slime", Quantity: math.MaxUint32},
		//{Type: "craft", Code: "iron", Quantity: math.MaxUint32},
		//{Type: "idle"},
		//{Type: "craft", Code: "iron_boots", Quantity: 5},
		//{Type: "task"},
		//{Type: "craft", Code: "fire_bow", Quantity: 5},
		//{Type: "harvest", Code: "spruce_tree", Quantity: math.MaxUint32},
		//{Type: "craft", Code: "copper", Quantity: math.MaxUint32},
		//{Type: "harvest", Code: "copper_rocks", Quantity: math.MaxUint32},
		//{Type: "harvest", Code: "iron_rocks", Quantity: math.MaxUint32},
	}

	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	var gridColumns []interface{}
	for i, charName := range characterNames {
		var rowWidgets []interface{}

		character := widgets.NewParagraph()
		character.Title = "Character"
		character.Text = charName

		rowWidgets = append(rowWidgets, character)

		action := widgets.NewParagraph()
		action.Title = "Action"
		action.Text = fmt.Sprintf("%s %s", tasks[i].Type, tasks[i].Code)

		rowWidgets = append(rowWidgets, action)

		status := widgets.NewParagraph()
		status.Title = "Status"
		status.Text = "Starting up"
		status.TextStyle.Fg = ui.ColorMagenta

		rowWidgets = append(rowWidgets, status)

		task := widgets.NewParagraph()
		task.Title = "Task"
		task.Text = ""

		rowWidgets = append(rowWidgets, task)

		cooldown := widgets.NewGauge()
		cooldown.Title = "Cooldown"
		cooldown.BarColor = ui.ColorRed
		cooldown.BorderStyle.Fg = ui.ColorWhite
		cooldown.TitleStyle.Fg = ui.ColorRed

		rowWidgets = append(rowWidgets, cooldown)

		levels := map[string]*widgets.Gauge{}
		for j, levelName := range levelNames {
			level := widgets.NewGauge()
			level.Title = strings.Title(levelName)
			level.BarColor = colors[j]
			level.BorderStyle.Fg = ui.ColorWhite
			level.TitleStyle.Fg = colors[j]

			levels[levelName] = level

			rowWidgets = append(rowWidgets, level)
		}

		characterWidgets[charName] = &CharacterWidget{
			Status:   status,
			Task:     task,
			Cooldown: cooldown,
			Levels:   levels,
		}

		var rows []interface{}
		rowHeight := 1.0 / float64(len(rowWidgets))
		for _, rowWidget := range rowWidgets {
			rows = append(rows, ui.NewRow(rowHeight, rowWidget))
		}

		columnWidth := 1.0 / float64(len(characterNames))
		gridColumns = append(gridColumns, ui.NewCol(columnWidth, rows...))
	}

	grid.Set(ui.NewRow(1, gridColumns...))

	ui.Render(grid)

	bank := &Bank{}

	characters := map[string]*Character{}
	for _, name := range characterNames {
		char := NewCharacter(c, bank, name)
		_, err = char.Get(ctx)
		if err != nil {
			log.Fatal(err)
		}
		characters[name] = char
	}

	updateWidgets(characters, grid)

	characterActions := make(chan characterAction)

	game, err := NewGame(ctx, c, bank)
	if err != nil {
		log.Fatal(err)
	}

	for i, name := range characterNames {
		runner := NewRunner(game, characters[name], characterActions)

		go func(runner *Runner, task CharacterTask) {
			for {
				switch task.Type {
				case "idle":
					return
				case "harvest":
					err = runner.harvest(ctx, task.Code, task.Quantity)
				case "craft":
					err = runner.craft(ctx, task.Code, task.Quantity)
					if err == nil {
						runner.reportAction("Crafting finished")
						return
					}
					if errors.Is(err, ErrCraftingNoMoreMaterials) {
						runner.reportAction("No more crafting materials")
						return
					}
					if errors.Is(err, ErrBankInsufficientQuantity) {
						runner.reportAction("Bank has no more crafting materials")
						return
					}
				case "item":
					err = runner.acquireItem(ctx, task.Code, task.Quantity)
				case "task":
					err = runner.taskLoop(ctx, NoopLoopStopper)
				case "task_when_not_crafting":
					err = runner.taskWhenNotCrafting(ctx, task.Code)
				case "fight":
					err = runner.fightMonsterLoop(ctx, task.Code, task.Quantity)
				}

				if err != nil {
					// TODO: Fatal
					runner.Char.Logger.Println(err)
				}
			}
		}(runner, tasks[i])
	}

	renderTicker := time.NewTicker(2 * time.Second)
	defer renderTicker.Stop()
	uiEvents := ui.PollEvents()

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				grid.SetRect(0, 0, payload.Width, payload.Height)
				ui.Clear()
				ui.Render(grid)
			}
		case <-renderTicker.C:
			updateWidgets(characters, grid)
		case a := <-characterActions:
			//a.Character.Logger.Print(a.Action)
			widget := characterWidgets[a.Character.Name]
			widget.Status.Text = a.Action
			//ui.Render(grid)
			ui.Render(widget.Status)
		}
	}
}

func updateWidgets(characters map[string]*Character, grid *ui.Grid) {
	for name, char := range characters {
		charWidget := characterWidgets[name]

		c := char.Character

		charWidget.Task.Text = fmt.Sprintf("%d/%d %s", c.TaskProgress, c.TaskTotal, c.Task)

		expiresIn := time.Until(char.CooldownExpires)
		if expiresIn.Seconds() > 0 {
			percent := expiresIn.Seconds() / float64(char.CooldownDuration)
			charWidget.Cooldown.Percent = int(percent * 100)
			charWidget.Cooldown.Label = fmt.Sprintf("%0.2fs", float64(expiresIn.Milliseconds())/1000.0)
		} else {
			charWidget.Cooldown.Percent = 0
			charWidget.Cooldown.Label = "0s"
		}

		for levelName, level := range charWidget.Levels {
			percent := char.GetPercentToLevel(levelName) * 100
			level.Percent = int(percent)
			level.Label = fmt.Sprintf("%0.2f%% (%d/%d xp)", percent, char.GetXP(levelName), char.GetMaxXP(levelName))
			level.Title = fmt.Sprintf("%s (%d)", strings.Title(levelName), char.GetLevel(levelName))
		}
	}
	ui.Render(grid)
}

type Runner struct {
	Game    *Game
	Char    *Character
	actions chan<- characterAction
}

func NewRunner(game *Game, char *Character, actions chan<- characterAction) *Runner {
	return &Runner{
		Game:    game,
		Char:    char,
		actions: actions,
	}
}

func (r *Runner) reportAction(fmtStr string, args ...interface{}) {
	r.actions <- characterAction{
		Character: r.Char,
		Action:    fmt.Sprintf(fmtStr, args...),
	}
}

func (r *Runner) acquireFromInventoryAndBank(ctx context.Context, code string, wantQuantity int) error {
	// First, see if we have enough of these in inventory
	inventoryQuantity := r.Char.InventoryQuantity(code)
	if inventoryQuantity >= wantQuantity {
		return nil
	}

	// Then see if we have enough in the bank+inventory together
	bankItems := r.Game.Bank.AsMap()

	bankQuantity := bankItems[code]

	// If the bank has any of these, withdraw them
	if bankQuantity > 0 {
		withdrawQuantity := bankQuantity
		if inventoryQuantity+bankQuantity > wantQuantity {
			withdrawQuantity -= inventoryQuantity
		}
		if withdrawQuantity > wantQuantity {
			withdrawQuantity = wantQuantity
		}

		r.reportAction("Moving to bank")
		err := r.Char.MoveClosest(ctx, r.Game.GetBankLocations())
		if err != nil {
			return fmt.Errorf("moving to bank: %w", err)
		}

		r.reportAction("Withdrawing from bank")
		_, err = r.Char.WithdrawBank(ctx, code, withdrawQuantity)
		if err != nil {
			if !errIsBankItemNotFound(err) && errIsBankInsufficientQuantity(err) {
				return fmt.Errorf("withdrawing from bank: %w", err)
			}
		}
	}

	return nil
}

func (r *Runner) withdrawCraftingMaterialsFromBank(ctx context.Context, craftingItems []CraftingItem, recipeQuantity int) error {
	err := r.Char.MoveClosest(ctx, r.Game.GetBankLocations())
	if err != nil {
		return fmt.Errorf("moving to bank: %w", err)
	}

	err = r.Char.BankAll(ctx)
	if err != nil {
		return fmt.Errorf("banking all: %w", err)
	}

	time.Sleep(time.Second)

	for _, craftingItem := range craftingItems {
		_, err = r.Char.WithdrawBank(ctx, craftingItem.Code, recipeQuantity*craftingItem.Quantity)
		if err != nil {
			if errIsBankInsufficientQuantity(err) {
				return ErrBankInsufficientQuantity
			}
			return fmt.Errorf("withdrawing from bank: %w", err)
		}
	}

	return nil
}

func (r *Runner) NumberCanCraftInBankAndInventory(crafting ItemCrafting) int {
	inventory := r.Char.NumberCanCraftInInventory(crafting)
	bank := r.Game.NumberCanCraftInBank(crafting)

	return inventory + bank
}

func (r *Runner) craft(ctx context.Context, itemCode string, wantQuantity int) error {
	item := r.Game.Items[itemCode]

	if item.Crafting == nil {
		return fmt.Errorf("item cannot be crafted")
	}

	maxCraftableInventorySize := r.Char.Character.InventoryMaxItems / item.Crafting.TotalQuantity()

	for wantQuantity > 0 {
		canCraftInInventory := r.Char.NumberCanCraftInInventory(*item.Crafting)
		canCraftInBank := r.Game.NumberCanCraftInBank(*item.Crafting)
		if canCraftInInventory+canCraftInBank == 0 {
			return ErrCraftingNoMoreMaterials
		}

		numToCraft := 0
		if canCraftInInventory > 0 {
			numToCraft = min(wantQuantity, canCraftInInventory)
			numToCraft = min(numToCraft, maxCraftableInventorySize)
		} else {
			numToCraft = min(wantQuantity, canCraftInBank)
			numToCraft = min(numToCraft, maxCraftableInventorySize)
		}

		if canCraftInInventory == 0 {
			r.reportAction("Not enough crafting materials, withdrawing from bank")

			if err := r.withdrawCraftingMaterialsFromBank(ctx, item.Crafting.Items, numToCraft); err != nil {
				return fmt.Errorf("withdrawing crafting materials from bank: %w", err)
			}
		}

		r.Char.WaitForCooldown()
		r.reportAction("Moving to workshop")

		err := r.Char.MoveClosest(ctx, r.Game.GetWorkshopLocations(item.Crafting.Skill))
		if err != nil {
			return fmt.Errorf("moving to workshop")
		}

		r.Char.WaitForCooldown()
		r.reportAction("Crafting %s", itemCode)

		_, err = r.Char.Craft(ctx, itemCode, numToCraft)
		if err != nil {
			return fmt.Errorf("crafting: %w", err)
		}

		r.Char.WaitForCooldown()

		wantQuantity -= numToCraft
	}

	return nil
}

func (r *Runner) harvest(ctx context.Context, resourceCode string, wantQuantity int) error {
	var resource client.ResourceSchema
	for _, res := range r.Game.Resources {
		if res.Code == resourceCode {
			resource = res
			break
		}
	}

	currentWeapon := r.Char.Character.WeaponSlot
	var currentWeaponAttack int
	if currentWeapon != "" {
		currentWeaponAttack = r.Game.Items[currentWeapon].Attack[string(resource.Skill)]
	}

	resistances := map[string]int{string(resource.Skill): 0}
	bestBankItem, bestBankAttack := r.Game.BestUsableBankWeapon(r.Char.Character.Level, resistances)

	if bestBankAttack > currentWeaponAttack && currentWeapon != bestBankItem {
		r.reportAction("Moving to bank")
		err := r.Char.MoveClosest(ctx, r.Game.GetBankLocations())
		if err != nil {
			return fmt.Errorf("moving to bank: %w", err)
		}

		if currentWeapon != "" {
			err := r.Char.Unequip(ctx, client.UnequipSchemaSlotWeapon)
			if err != nil {
				return fmt.Errorf("unequipping slot: %w", err)
			}

			_, err = r.Char.DepositBank(ctx, currentWeapon, 1)
			if err != nil {
				return fmt.Errorf("depositing weapon in bank: %w", err)
			}
		}

		_, err = r.Char.WithdrawBank(ctx, bestBankItem, 1)
		if err != nil {
			return fmt.Errorf("withdrawing weapon from bank: %w", err)
		}

		err = r.Char.Equip(ctx, client.EquipSchemaSlotWeapon, bestBankItem)
		if err != nil {
			return fmt.Errorf("equipping slot: %w", err)
		}
	}

	for harvestedQuantity := 0; harvestedQuantity < wantQuantity; harvestedQuantity++ {
		err := r.BankIfInventoryFull(ctx)
		if err != nil {
			return err
		}

		r.reportAction("Harvesting %s", resourceCode)
		err = r.Char.MoveClosest(ctx, r.Game.GetResourceLocations(resourceCode))
		if err != nil {
			return err
		}

		r.Char.WaitForCooldown()

		_, err = r.Char.Gather(ctx)
		if err != nil {
			return fmt.Errorf("gathering: %w", err)
		}
	}

	return nil
}

func (r *Runner) acquireItem(ctx context.Context, itemCode string, wantQuantity int) error {
	r.reportAction("Acquiring %d %s", wantQuantity, itemCode)

	err := r.acquireFromInventoryAndBank(ctx, itemCode, wantQuantity)
	if err != nil {
		return err
	}

	// Check inventory to see if we have enough
	inventoryQuantity := r.Char.InventoryQuantity(itemCode)
	if inventoryQuantity >= wantQuantity {
		return nil
	}

	item := r.Game.Items[itemCode]
	requiredSkills := item.GetAllSkillRequirements()

	for skill, requiredLevel := range requiredSkills {
		err := r.TrainSkill(ctx, skill, requiredLevel)
		if err != nil {
			return fmt.Errorf("training: %w", err)
		}
	}

	// Reduce the desired quantity by the quantity we already have
	//wantQuantity -= inventoryQuantity

	// TODO: It doesn't really make sense to withdraw from the bank right before we start training.
	//  Only withdraw from the bank after training, or if we don't need training

	canBeCrafted := item.Crafting != nil
	if canBeCrafted {
		return r.craft(ctx, itemCode, wantQuantity)
	} else {
		var selectedResource *client.ResourceSchema
		for _, drop := range r.Game.Drops[item.Code] {
			if drop.Level > r.Char.GetLevel(string(drop.Skill)) {
				continue
			}
			selectedResource = &drop
			break
		}

		return r.harvest(ctx, selectedResource.Code, wantQuantity)
	}
}

func (r *Runner) BankIfInventoryFull(ctx context.Context) error {
	if !r.Char.IsInventoryFull() {
		return nil
	}

	err := r.MoveToBankAndDepositAll(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *Runner) MoveToBankAndDepositAll(ctx context.Context) error {
	r.reportAction("Moving to bank")
	err := r.Char.MoveClosest(ctx, r.Game.GetBankLocations())
	if err != nil {
		return fmt.Errorf("moving to bank: %w", err)
	}

	r.reportAction("Depositing all in bank")
	err = r.Char.BankAll(ctx)
	if err != nil {
		return fmt.Errorf("depositing all in bank: %w", err)
	}
	return nil
}

func (r *Runner) taskWhenNotCrafting(ctx context.Context, itemCode string) error {
	item := r.Game.Items[itemCode]

	if item.Crafting == nil {
		return fmt.Errorf("item cannot be crafted")
	}

	recipeRequiredInvSpace := item.Crafting.TotalQuantity()

	canMakeInventoryOfItem := func() bool {
		numCanCraft := r.NumberCanCraftInBankAndInventory(*item.Crafting)

		// Only stop when we can make a full inventory worth
		fullInvQuantity := r.Char.Character.InventoryMaxItems / recipeRequiredInvSpace
		return numCanCraft >= fullInvQuantity
	}

	for {
		err := r.taskLoop(ctx, canMakeInventoryOfItem)
		if err != nil {
			return fmt.Errorf("task loop: %w", err)
		}

		err = r.MoveToBankAndDepositAll(ctx)
		if err != nil {
			return err
		}

		numCanCraft := r.NumberCanCraftInBankAndInventory(*item.Crafting)

		// Only make whole inventories
		fullInvQuantity := r.Char.Character.InventoryMaxItems / recipeRequiredInvSpace
		numInventories := numCanCraft / fullInvQuantity
		numToCraft := numInventories * fullInvQuantity

		err = r.craft(ctx, itemCode, numToCraft)
		if err != nil {
			return fmt.Errorf("crafting: %w", err)
		}
	}
}
