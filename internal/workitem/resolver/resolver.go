// Package resolver provides functionality to build hierarchical trees of work items.
package resolver

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"sync"

	"github.com/contextvibes/cli/internal/workitem"
)

// issueLinkRegex finds GitHub task list items like '- [ ] #123' or '- [x] #456'.
var issueLinkRegex = regexp.MustCompile(`-\s+\[\s*[xX]?\s*]\s+#(\d+)`)

// HierarchyResolver builds a tree of work items based on task list relationships.
type HierarchyResolver struct {
	provider workitem.Provider
}

// New creates a new HierarchyResolver.
func New(provider workitem.Provider) *HierarchyResolver {
	return &HierarchyResolver{
		provider: provider,
	}
}

// BuildTree recursively fetches and assembles a work item and its children.
func (r *HierarchyResolver) BuildTree(
	ctx context.Context,
	rootItemNumber int,
	withComments bool,
) (*workitem.WorkItem, error) {
	// Fetch the root item, passing the withComments flag down.
	rootItem, err := r.provider.GetItem(ctx, rootItemNumber, withComments)
	if err != nil {
		return nil, fmt.Errorf("could not fetch root item #%d: %w", rootItemNumber, err)
	}

	childNumbers := r.parseChildIssueNumbers(rootItem.Body)
	if len(childNumbers) == 0 {
		return rootItem, nil // This item is a leaf node
	}

	var waitGroup sync.WaitGroup

	childChan := make(chan *workitem.WorkItem, len(childNumbers))
	errChan := make(chan error, len(childNumbers))

	for _, childNumber := range childNumbers {
		waitGroup.Add(1)

		go func(num int) {
			defer waitGroup.Done()
			// Recursively call BuildTree, passing the withComments flag through.
			childTree, err := r.BuildTree(ctx, num, withComments)
			if err != nil {
				errChan <- err

				return
			}

			childChan <- childTree
		}(childNumber)
	}

	waitGroup.Wait()
	close(childChan)
	close(errChan)

	for err := range errChan {
		// For now, we'll just return the first error we encounter.
		return nil, err
	}

	for child := range childChan {
		rootItem.Children = append(rootItem.Children, child)
	}

	return rootItem, nil
}

// parseChildIssueNumbers extracts all issue numbers from GitHub task lists in a string.
func (r *HierarchyResolver) parseChildIssueNumbers(body string) []int {
	matches := issueLinkRegex.FindAllStringSubmatch(body, -1)
	if matches == nil {
		return nil
	}

	numbers := make([]int, 0, len(matches))
	for _, match := range matches {
		//nolint:mnd // Regex match count 2 is specific to this pattern.
		if len(match) == 2 {
			//nolint:noinlineerr // Inline check is standard for Atoi.
			if num, err := strconv.Atoi(match[1]); err == nil {
				numbers = append(numbers, num)
			}
		}
	}

	return numbers
}
