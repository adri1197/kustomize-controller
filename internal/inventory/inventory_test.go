/*
Copyright 2022 The Flux authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package inventory

import (
	"os"
	"strings"
	"testing"

	"github.com/fluxcd/pkg/ssa"
	ssautil "github.com/fluxcd/pkg/ssa/utils"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/fluxcd/cli-utils/pkg/object"
)

func Test_Inventory(t *testing.T) {
	g := NewWithT(t)

	set1, err := readManifest("testdata/inventory1.yaml")
	if err != nil {
		t.Fatal(err)
	}

	inv1 := New()
	err = AddChangeSet(inv1, set1)
	g.Expect(err).ToNot(HaveOccurred())

	set2, err := readManifest("testdata/inventory2.yaml")
	if err != nil {
		t.Fatal(err)
	}

	inv2 := New()
	err = AddChangeSet(inv2, set2)
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("lists objects in inventory", func(t *testing.T) {
		unList, err := List(inv1)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(unList)).To(BeIdenticalTo(len(inv1.Entries)))

		mList, err := ListMetadata(inv1)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(mList)).To(BeIdenticalTo(len(inv1.Entries)))
	})

	t.Run("diff objects in inventory", func(t *testing.T) {
		unList, err := Diff(inv2, inv1, nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(unList)).To(BeIdenticalTo(1))
		g.Expect(unList[0].GetName()).To(BeIdenticalTo("test2"))
	})

	t.Run("diff objects in inventory ignoring skipped", func(t *testing.T) {
		skipped := object.ObjMetadata{
			Name:      "test2",
			Namespace: "test",
			GroupKind: schema.GroupKind{
				Group: "",
				Kind:  "ConfigMap",
			},
		}
		unList, err := Diff(inv2, inv1, map[object.ObjMetadata]struct{}{skipped: {}})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(unList)).To(BeIdenticalTo(0))
	})
}

func readManifest(manifest string) (*ssa.ChangeSet, error) {
	data, err := os.ReadFile(manifest)
	if err != nil {
		return nil, err
	}

	objects, err := ssautil.ReadObjects(strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}

	cs := ssa.NewChangeSet()

	for _, o := range objects {
		cse := ssa.ChangeSetEntry{
			ObjMetadata:  object.UnstructuredToObjMetadata(o),
			GroupVersion: o.GroupVersionKind().Version,
			Subject:      ssautil.FmtUnstructured(o),
			Action:       ssa.CreatedAction,
		}
		cs.Add(cse)
	}

	return cs, nil
}
