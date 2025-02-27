/*
Copyright 2021 The KubeVela Authors.

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

package usecase

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/oam-dev/kubevela/apis/core.oam.dev/v1beta1"
	"github.com/oam-dev/kubevela/apis/types"
	v1 "github.com/oam-dev/kubevela/pkg/apiserver/rest/apis/v1"
	"github.com/oam-dev/kubevela/pkg/apiserver/rest/utils"
	"github.com/oam-dev/kubevela/pkg/oam/util"
)

var _ = Describe("Test namespace usecase functions", func() {
	var (
		definitionUsecase *definitionUsecaseImpl
	)

	BeforeEach(func() {
		definitionUsecase = &definitionUsecaseImpl{kubeClient: k8sClient, caches: make(map[string]*utils.MemoryCache)}
		err := k8sClient.Create(context.Background(), &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "vela-system",
			},
		})
		Expect(err).Should(SatisfyAny(BeNil(), &util.AlreadyExistMatcher{}))
	})
	It("Test ListDefinitions function", func() {
		By("List component definitions")
		webserver, err := ioutil.ReadFile("./testdata/webserver-cd.yaml")
		Expect(err).Should(Succeed())
		var cd v1beta1.ComponentDefinition
		err = yaml.Unmarshal(webserver, &cd)
		Expect(err).Should(Succeed())
		err = k8sClient.Create(context.Background(), &cd)
		Expect(err).Should(Succeed())
		components, err := definitionUsecase.ListDefinitions(context.TODO(), "", "component")
		Expect(err).Should(BeNil())
		Expect(cmp.Diff(len(components), 1)).Should(BeEmpty())
		Expect(cmp.Diff(components[0].Name, "webservice-test")).Should(BeEmpty())
		Expect(components[0].Description).ShouldNot(BeEmpty())

		By("List trait definitions")
		myingress, err := ioutil.ReadFile("./testdata/myingress-td.yaml")
		Expect(err).Should(Succeed())
		var td v1beta1.TraitDefinition
		err = yaml.Unmarshal(myingress, &td)
		Expect(err).Should(Succeed())
		err = k8sClient.Create(context.Background(), &td)
		Expect(err).Should(Succeed())
		traits, err := definitionUsecase.ListDefinitions(context.TODO(), "", "trait")
		Expect(err).Should(BeNil())
		Expect(cmp.Diff(len(traits), 1)).Should(BeEmpty())
		Expect(cmp.Diff(traits[0].Name, "myingress")).Should(BeEmpty())
		Expect(traits[0].Description).ShouldNot(BeEmpty())

		By("List workflow step definitions")
		step, err := ioutil.ReadFile("./testdata/applyapplication-sd.yaml")
		Expect(err).Should(Succeed())
		var sd v1beta1.WorkflowStepDefinition
		err = yaml.Unmarshal(step, &sd)
		Expect(err).Should(Succeed())
		err = k8sClient.Create(context.Background(), &sd)
		Expect(err).Should(Succeed())
		wfstep, err := definitionUsecase.ListDefinitions(context.TODO(), "", "workflowstep")
		Expect(err).Should(BeNil())
		Expect(cmp.Diff(len(wfstep), 1)).Should(BeEmpty())
		Expect(cmp.Diff(wfstep[0].Name, "apply-application")).Should(BeEmpty())
		Expect(wfstep[0].Description).ShouldNot(BeEmpty())
	})

	It("Test DetailDefinition function", func() {
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "workflowstep-schema-apply-object",
				Namespace: "vela-system",
			},
			Data: map[string]string{
				types.OpenapiV3JSONSchema: `{"properties":{"batchPartition":{"title":"batchPartition","type":"integer"},"volumes": {"description":"Specify volume type, options: pvc, configMap, secret, emptyDir","enum":["pvc","configMap","secret","emptyDir"],"title":"volumes","type":"string"}, "rolloutBatches":{"items":{"properties":{"replicas":{"title":"replicas","type":"integer"}},"required":["replicas"],"type":"object"},"title":"rolloutBatches","type":"array"},"targetRevision":{"title":"targetRevision","type":"string"},"targetSize":{"title":"targetSize","type":"integer"}},"required":["targetRevision","targetSize"],"type":"object"}`,
			},
		}
		err := k8sClient.Create(context.Background(), cm)
		Expect(err).Should(Succeed())
		schema, err := definitionUsecase.DetailDefinition(context.TODO(), "apply-object", "workflowstep")
		Expect(err).Should(Succeed())

		schemaFromCM := &openapi3.Schema{}
		err = schemaFromCM.UnmarshalJSON([]byte(cm.Data["openapi-v3-json-schema"]))
		Expect(err).Should(Succeed())

		Expect(schema.APISchema).Should(Equal(schemaFromCM))
	})

	It("Test renderDefaultUISchema", func() {
		schema := &v1.DetailDefinitionResponse{}
		data, err := ioutil.ReadFile("./testdata/api-schema.json")
		Expect(err).Should(Succeed())
		err = json.Unmarshal(data, schema)
		Expect(err).Should(Succeed())
		Expect(cmp.Diff(len(schema.APISchema.Required), 3)).Should(BeEmpty())
		uiSchema := renderDefaultUISchema(schema.APISchema)
		Expect(cmp.Diff(len(uiSchema), 12)).Should(BeEmpty())
	})

	It("Test patchSchema", func() {
		ddr := &v1.DetailDefinitionResponse{}
		data, err := ioutil.ReadFile("./testdata/api-schema.json")
		Expect(err).Should(Succeed())
		err = json.Unmarshal(data, ddr)
		Expect(err).Should(Succeed())
		Expect(cmp.Diff(len(ddr.APISchema.Required), 3)).Should(BeEmpty())
		defaultschema := renderDefaultUISchema(ddr.APISchema)

		customschema := []*utils.UIParameter{}
		cdata, err := ioutil.ReadFile("./testdata/ui-custom-schema.yaml")
		Expect(err).Should(Succeed())
		err = yaml.Unmarshal(cdata, &customschema)
		Expect(err).Should(Succeed())

		uiSchema := patchSchema(defaultschema, customschema)
		Expect(cmp.Diff(len(uiSchema), 12)).Should(BeEmpty())
		Expect(cmp.Diff(uiSchema[7].JSONKey, "livenessProbe")).Should(BeEmpty())
		Expect(cmp.Diff(len(uiSchema[7].SubParameters), 8)).Should(BeEmpty())

		outdata, err := yaml.Marshal(uiSchema)
		Expect(err).Should(Succeed())
		err = ioutil.WriteFile("./testdata/ui-schema.yaml", outdata, 0755)
		Expect(err).Should(Succeed())
	})
})

func TestAddDefinitionUISchema(t *testing.T) {
	du := NewDefinitionUsecase()
	cdata, err := ioutil.ReadFile("./testdata/ui-custom-schema.yaml")
	if err != nil {
		t.Fatal(err)
	}
	_, err = du.AddDefinitionUISchema(context.TODO(), "webservice", "component", string(cdata))
	if err != nil {
		t.Fatal(err)
	}
}
