package functional

import (
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
)

const (
	fluentForwardOutputName        = "fluentForward"
	elasticSearchForwardOutputName = "elasticSearchForward"
	forwardPipelineName            = "forward-pipeline"
)

type ClusterLogForwarderBuilder struct {
	Forwarder *logging.ClusterLogForwarder
}

type PipelineBuilder struct {
	clfb      *ClusterLogForwarderBuilder
	inputName string
}

type OutputSpecVisiter func(spec *logging.OutputSpec)

func NewClusterLogForwarderBuilder(clf *logging.ClusterLogForwarder) *ClusterLogForwarderBuilder {
	return &ClusterLogForwarderBuilder{
		Forwarder: clf,
	}
}

func (b *ClusterLogForwarderBuilder) FromInput(inputName string) *PipelineBuilder {
	pipelineBuilder := &PipelineBuilder{
		clfb:      b,
		inputName: inputName,
	}
	return pipelineBuilder
}

func (p *PipelineBuilder) ToFluentForwardOutput() *ClusterLogForwarderBuilder {
	return p.ToOutputWithVisitor(func(output *logging.OutputSpec) {}, fluentForwardOutputName)
}
func (p *PipelineBuilder) ToElasticSearchOutput() *ClusterLogForwarderBuilder {
	return p.ToOutputWithVisitor(func(output *logging.OutputSpec) {}, elasticSearchForwardOutputName)
}

func (p *PipelineBuilder) ToOutputWithVisitor(visit OutputSpecVisiter, forwardOutputName string) *ClusterLogForwarderBuilder {
	clf := p.clfb.Forwarder
	outputs := clf.Spec.OutputMap()
	var output *logging.OutputSpec
	var found bool
	if output, found = outputs[fluentForwardOutputName]; !found {
		if forwardOutputName == fluentForwardOutputName {
			output = &logging.OutputSpec{
				Name: fluentForwardOutputName,
				Type: logging.OutputTypeFluentdForward,
				URL:  "tcp://0.0.0.0:24224",
			}
		} else if forwardOutputName == elasticSearchForwardOutputName {
			output = &logging.OutputSpec{
				Name: elasticSearchForwardOutputName,
				Type: logging.OutputTypeElasticsearch,
				URL:  "https://0.0.0.0:9200",
			}
		}

		visit(output)
		clf.Spec.Outputs = append(clf.Spec.Outputs, *output)
	}
	added := false
	clf.Spec.Pipelines, added = addInputToPipeline(p.inputName, forwardPipelineName, clf.Spec.Pipelines)
	if !added {
		clf.Spec.Pipelines = append(clf.Spec.Pipelines, logging.PipelineSpec{
			Name:       forwardPipelineName,
			InputRefs:  []string{p.inputName},
			OutputRefs: []string{output.Name},
		})
	}
	return p.clfb
}

func addInputToPipeline(inputName, pipelineName string, pipelineSpecs []logging.PipelineSpec) ([]logging.PipelineSpec, bool) {
	pipelines := []logging.PipelineSpec{}
	found := false
	for _, pipeline := range pipelineSpecs {
		if pipelineName == pipeline.Name {
			found = true
			pipeline.InputRefs = append(pipeline.InputRefs, inputName)
		}
		pipelines = append(pipelines, pipeline)
	}
	return pipelines, found
}
