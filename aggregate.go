package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/beevik/etree"
)

type aggregator struct {
	XMLName   xml.Name `xml:"EntitiesDescriptor"`
	Entities  map[string]*etree.Element
	rootAttrs map[string]etree.Attr
}

func (a *aggregator) add(r io.Reader) error {

	if a.Entities == nil {
		a.Entities = map[string]*etree.Element{}
		a.rootAttrs = map[string]etree.Attr{}
	}

	d := etree.NewDocument()
	if _, err := d.ReadFrom(r); err != nil {
		return err
	}

	root := d.Root()

	if root == nil {
		return fmt.Errorf("Nil root element")
	}

	for _, attr := range root.Attr {
		a.rootAttrs[strings.Join([]string{attr.Space, attr.Key}, ":")] = attr
	}

	fmt.Printf("root = %+v\n", root)

	switch root.Tag {
	case "EntitiesDescriptor":
		for _, child := range root.ChildElements() {
			fmt.Printf("child = %+v\n", child)
			a.addElement(child)
		}
	case "EntityDescriptor":
		a.addElement(root)
	}

	return nil
}

func (a *aggregator) addElement(e *etree.Element) {
	k := e.SelectAttrValue("entityID", "")
	if k != "" {
		a.Entities[k] = e
	}
}

func (a *aggregator) Doc() *etree.Document {
	d := etree.NewDocument()
	ele := etree.NewElement("EntitiesDescriptor")
	ele.Attr = make([]etree.Attr, 0, len(a.rootAttrs))
	for _, attr := range a.rootAttrs {
		ele.Attr = append(ele.Attr, attr)
	}

	for _, entity := range a.Entities {
		ele.AddChild(entity)
	}

	d.SetRoot(ele)

	return d
}

func (a *aggregator) WriteTo(w io.Writer) (int64, error) {
	return a.Doc().WriteTo(w)
}
