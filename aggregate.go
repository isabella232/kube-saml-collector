package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/beevik/etree"
	"github.com/golang/glog"
)

type aggregator struct {
	XMLName  xml.Name `xml:"EntitiesDescriptor"`
	Entities map[string]*etree.Element
	xmlns    map[string]etree.Attr
}

const (
	metadataNamespaceURN = "urn:oasis:names:tc:SAML:2.0:metadata"

	xmlns = "xmlns"
)

func (a *aggregator) init() {
	if a.Entities == nil {
		a.Entities = map[string]*etree.Element{}
	}
	if a.xmlns == nil {
		a.xmlns = map[string]etree.Attr{}
	}
}

func (a *aggregator) add(r io.Reader) error {
	a.init()

	d := etree.NewDocument()
	if _, err := d.ReadFrom(r); err != nil {
		return err
	}

	root := d.Root()

	if root == nil {
		return fmt.Errorf("Nil root element")
	}

	for _, attr := range root.Attr {
		if attr.Key == xmlns || attr.Space == xmlns {
			a.xmlns[strings.Join([]string{attr.Space, attr.Key}, ":")] = attr
		}
	}

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
	a.init()

	k := e.SelectAttrValue("entityID", "")
	glog.Infof("Adding entity id \"%s\".\n", k)
	if k != "" {
		a.Entities[k] = e
	}
}

func (a *aggregator) Doc() *etree.Document {
	a.init()
	a.xmlns[":xmlns"] = etree.Attr{Key: xmlns, Value: metadataNamespaceURN}

	d := etree.NewDocument()
	ele := etree.NewElement("EntitiesDescriptor")
	ele.Attr = make([]etree.Attr, 0, len(a.xmlns))

	for _, attr := range a.xmlns {
		ele.Attr = append(ele.Attr, attr)
	}

	for _, entity := range a.Entities {
		ele.AddChild(entity)
	}

	d.SetRoot(ele)

	return d
}

func (a *aggregator) WriteTo(w io.Writer) (int64, error) {
	a.init()
	return a.Doc().WriteTo(w)
}
