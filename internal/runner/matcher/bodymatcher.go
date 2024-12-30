package matcher

import (
	"context"
	"fmt"

	"github.com/ablankz/bloader/internal/container"
	"github.com/ablankz/bloader/internal/logger"
)

// BodyCondition represents the body condition
type BodyCondition struct {
	ID        *string        `yaml:"id"`
	Extractor *DataExtractor `yaml:"extractor"`
}

// BodyConditionMatcher represents the body matcher
type BodyConditionMatcher func(body any) (bool, error)

// MatcherGenerate generates the body matcher
func (bc BodyCondition) MatcherGenerate(ctx context.Context, ctr *container.Container) (BodyConditionMatcher, error) {
	if bc.ID == nil {
		return nil, fmt.Errorf("id is required")
	}
	if bc.Extractor == nil {
		return nil, fmt.Errorf("extractor is required")
	}
	extractor, err := bc.Extractor.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed to validate extractor: %v", err)
	}
	return func(body any) (bool, error) {
		res, err := extractor.Extract(body)
		if err != nil {
			return false, fmt.Errorf("failed to extract body: %v", err)
		}
		var match bool
		if v, ok := res.(bool); ok {
			if v {
				match = true
			}
		} else {
			ctr.Logger.Warn(ctx, "The result of the jmespath query is not a boolean",
				logger.Value("on", "runResponseHandler"))
		}
		return match, nil
	}, nil
}

// BodyConditions represents a slice of BodyCondition
type BodyConditions []BodyCondition

// BodyConditionsMatcher represents the body matcher
type BodyConditionsMatcher func(body any) (string, bool, error)

// MatcherGenerate generates the body matcher
func (bcs BodyConditions) MatcherGenerate(ctx context.Context, ctr *container.Container) (BodyConditionsMatcher, error) {
	var matchers []BodyConditionMatcher
	for _, bc := range bcs {
		matcher, err := bc.MatcherGenerate(ctx, ctr)
		if err != nil {
			return nil, fmt.Errorf("failed to generate body matcher: %v", err)
		}
		matchers = append(matchers, matcher)
	}
	return func(body any) (string, bool, error) {
		for i, matcher := range matchers {
			match, err := matcher(body)
			if err != nil {
				return *bcs[i].ID, false, fmt.Errorf("failed to match body: %v", err)
			}
			if match {
				return *bcs[i].ID, true, nil
			}
		}
		return "", false, nil
	}, nil
}