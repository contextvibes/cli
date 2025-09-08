// cmd/project/issues/internal/display.go
package internal

import (
	"fmt"
	"strings"

	"github.com/contextvibes/cli/internal/ui"
	"github.com/contextvibes/cli/internal/workitem"
)

// DisplayWorkItem renders a detailed, multi-line view of a single work item.
// It does not handle comments, which are displayed by the calling command.
func DisplayWorkItem(p *ui.Presenter, item *workitem.WorkItem) {
	p.Header(fmt.Sprintf("%s (#%d)", item.Title, item.Number))
	p.Detail("State: %s, Author: %s, Created: %s", item.State, item.Author, item.CreatedAt.Format("2006-01-02"))

	if len(item.Labels) > 0 {
		p.Detail("Labels: %s", strings.Join(item.Labels, ", "))
	}
	if len(item.Assignees) > 0 {
		p.Detail("Assignees: %s", strings.Join(item.Assignees, ", "))
	}

	p.Separator()
	fmt.Fprintln(p.Out(), item.Body)
	p.Separator()
}
