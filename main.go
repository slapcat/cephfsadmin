package main

import (
	"encoding/json"
	"fmt"
	"github.com/ceph/go-ceph/cephfs/admin"
	"github.com/ceph/go-ceph/rados"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strings"
)

var subvol []string

type Ref struct {
	vol         string
	subvolgroup string
	subvol      string
}

func main() {

	conn, _ := rados.NewConn()
	conn.ReadDefaultConfigFile()
	conn.Connect()

	fsa := admin.NewFromConn(conn)
	fsid, err := conn.GetFSID()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	vol, err := fsa.ListVolumes()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	flex := tview.NewFlex()

	textArea := tview.NewTextArea().
		SetWrap(true).
		SetPlaceholder("")
	textArea.SetTitle("Details").SetBorder(true)

	ctlArea := tview.NewTextArea().SetWrap(true)
	ctlArea.SetText("v: create volume   G: create subvolumegroup  s: create subvolume  S: manage snapshots  D: delete  m: edit metadata", true)

	form := tview.NewForm().
		AddDropDown("Title", []string{"Mr.", "Ms.", "Mrs.", "Dr.", "Prof."}, 0, nil).
		AddInputField("First name", "", 20, nil, nil).
		AddInputField("Last name", "", 20, nil, nil).
		AddTextArea("Address", "", 40, 0, 0, nil).
		AddTextView("Notes", "This is just a demo.\nYou can enter whatever you wish.", 40, 2, true, false).
		AddCheckbox("Age 18+", false, nil).
		AddPasswordField("Password", "", 10, '*', nil).
		AddButton("Save", nil).
		AddButton("Quit", func() {
			// nothing
		})
	form.SetBorder(true).SetTitle("Enter some data").SetTitleAlign(tview.AlignLeft)

	var tree = tview.NewTreeView()
	tree.SetTitle("CephFS Volumes").SetBorder(true)

	root := tview.NewTreeNode(fsid)
	for _, v := range vol {
		root.AddChild(tview.NewTreeNode(v))
	}

	tree.SetRoot(root).
		SetCurrentNode(root).
		SetSelectedFunc(func(n *tview.TreeNode) {
			r := Ref{}

			if n.GetLevel() == 1 && n.GetChildren() == nil {
				subvolgroup, err := fsa.ListSubVolumeGroups(n.GetText())
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				}

				r.vol = n.GetText()
				r.subvolgroup = ""
				n.AddChild(tview.NewTreeNode("_nogroup").SetReference(r))

				for _, g := range subvolgroup {
					r.subvolgroup = g
					n.AddChild(tview.NewTreeNode(g).SetReference(r))
				}

			}

			if n.GetLevel() == 2 && n.GetChildren() == nil {

				r = n.GetReference().(Ref)

				if n.GetText() == "_nogroup" {
					r.subvolgroup = ""
					subvol, err = fsa.ListSubVolumes(r.vol, "")
					if err != nil {
						fmt.Printf("Error: %v\n", err)
					}
				} else {
					r.subvolgroup = n.GetText()
					subvol, err = fsa.ListSubVolumes(r.vol, n.GetText())
					if err != nil {

						fmt.Printf("Error: %v\n", err)
					}
				}

				for _, s := range subvol {
					n.AddChild(tview.NewTreeNode(s).SetReference(r))
				}
			}

			if n.GetLevel() == 3 {
				r = n.GetReference().(Ref)
				info, err := fsa.SubVolumeInfo(r.vol, r.subvolgroup, n.GetText())
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				}

				infoText := prettyPrint(info)
				textArea.SetText(infoText, true)
			}

		}).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Rune() == 'G' {
				flex.AddItem(form, 0, 2, true)
			}
			return event
		})

	flex = tview.NewFlex().
		AddItem(tree, 0, 1, true).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(textArea, 0, 2, false).
			AddItem(ctlArea, 2, 1, false), 0, 2, false)

	tview.NewApplication().
		SetRoot(flex, true).
		Run()

		//does this work? YES!!!
	fmt.Println(tree.GetCurrentNode().GetText())

	conn.Shutdown()

}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func spew(ref map[string]string, textArea *tview.TextArea) {

	var settings []string = nil
	for k, v := range ref {

		settings = append(settings, fmt.Sprintf("%k : %v", k, v))

	}

	textArea.SetText(strings.Join(settings, " | "), true)

}
