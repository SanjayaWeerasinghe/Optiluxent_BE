package mmcategories

import (
	"context"
	"fmt"
	"strings"
)

type Service struct {
	repo     Repository
	maxDepth int
}

func NewService(repo Repository, maxDepth int) *Service {
	return &Service{repo: repo, maxDepth: maxDepth}
}

func (s *Service) MaxDepth() int { return s.maxDepth }

func (s *Service) Count(ctx context.Context, tenantID uint) (int64, error) {
	return s.repo.Count(ctx, tenantID)
}

func (s *Service) List(ctx context.Context, tenantID uint, limit, offset int) ([]CategoryResponse, error) {
	// Fetch all rows to build the lookup map for path enrichment
	allRows, err := s.repo.List(ctx, tenantID, 0, 0)
	if err != nil {
		return nil, err
	}
	byID := make(map[uint]*MaterialCategory, len(allRows))
	for i := range allRows {
		byID[allRows[i].ID] = &allRows[i]
	}
	// Fetch only the requested page
	rows, err := s.repo.List(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]CategoryResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, s.buildResponse(r, byID))
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, tenantID, id uint) (*CategoryResponse, error) {
	c, err := s.repo.Get(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("category not found")
	}
	all, err := s.repo.List(ctx, tenantID, 0, 0)
	if err != nil {
		return nil, err
	}
	byID := make(map[uint]*MaterialCategory, len(all))
	for i := range all {
		byID[all[i].ID] = &all[i]
	}
	resp := s.buildResponse(*c, byID)
	return &resp, nil
}

func (s *Service) Create(ctx context.Context, tenantID uint, req *CreateCategoryRequest) (*CategoryResponse, error) {
	if req.ParentID != nil {
		depth, err := s.computeDepth(ctx, tenantID, *req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent category not found")
		}
		if depth+1 >= s.maxDepth {
			return nil, fmt.Errorf("maximum category depth of %d exceeded", s.maxDepth)
		}
	}
	c := &MaterialCategory{
		TenantID: tenantID,
		ParentID: req.ParentID,
		Code:     req.Code,
		Name:     req.Name,
		IsActive: true,
	}
	if err := s.repo.Create(ctx, c); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, fmt.Errorf("category code '%s' already exists", req.Code)
		}
		return nil, err
	}
	return s.Get(ctx, tenantID, c.ID)
}

func (s *Service) Update(ctx context.Context, tenantID, id uint, req *UpdateCategoryRequest) (*CategoryResponse, error) {
	c, err := s.repo.Get(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("category not found")
	}
	if req.ParentID != nil {
		// Prevent setting parent to self or a descendant
		if *req.ParentID == id {
			return nil, fmt.Errorf("a category cannot be its own parent")
		}
		if err := s.checkNotDescendant(ctx, tenantID, id, *req.ParentID); err != nil {
			return nil, err
		}
		depth, err := s.computeDepth(ctx, tenantID, *req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent category not found")
		}
		if depth+1 >= s.maxDepth {
			return nil, fmt.Errorf("maximum category depth of %d exceeded", s.maxDepth)
		}
		c.ParentID = req.ParentID
	} else if req.ParentID == nil && c.ParentID != nil {
		// Explicit null means move to root — only if the field is present; we use pointer so nil
		// means "not provided". Since UpdateCategoryRequest.ParentID is *uint, we can't
		// distinguish "omitted" from "set to null". Convention: if nil is sent, move to root.
		c.ParentID = nil
	}
	if req.Name != "" {
		c.Name = req.Name
	}
	if req.IsActive != nil {
		c.IsActive = *req.IsActive
	}
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	return s.Get(ctx, tenantID, id)
}

func (s *Service) Delete(ctx context.Context, tenantID, id uint) error {
	if _, err := s.repo.Get(ctx, tenantID, id); err != nil {
		return fmt.Errorf("category not found")
	}
	return s.repo.Delete(ctx, tenantID, id)
}

// computeDepth walks the ancestor chain (depth of the given node, root = 0).
func (s *Service) computeDepth(ctx context.Context, tenantID, id uint) (int, error) {
	depth := 0
	cur := id
	seen := map[uint]bool{}
	for {
		if seen[cur] {
			return 0, fmt.Errorf("circular reference detected")
		}
		seen[cur] = true
		c, err := s.repo.Get(ctx, tenantID, cur)
		if err != nil {
			return 0, err
		}
		if c.ParentID == nil {
			return depth, nil
		}
		cur = *c.ParentID
		depth++
	}
}

// checkNotDescendant ensures targetID is not a descendant of sourceID.
func (s *Service) checkNotDescendant(ctx context.Context, tenantID, sourceID, targetID uint) error {
	cur := targetID
	seen := map[uint]bool{}
	for {
		if cur == sourceID {
			return fmt.Errorf("cannot set a descendant as parent")
		}
		if seen[cur] {
			return nil
		}
		seen[cur] = true
		c, err := s.repo.Get(ctx, tenantID, cur)
		if err != nil {
			return nil
		}
		if c.ParentID == nil {
			return nil
		}
		cur = *c.ParentID
	}
}

func (s *Service) buildResponse(c MaterialCategory, byID map[uint]*MaterialCategory) CategoryResponse {
	resp := CategoryResponse{MaterialCategory: c}

	// Build path and depth by walking ancestors
	segments := []string{c.Name}
	depth := 0
	cur := c.ParentID
	for cur != nil {
		p, ok := byID[*cur]
		if !ok {
			break
		}
		segments = append([]string{p.Name}, segments...)
		resp.ParentCode = p.Code
		resp.ParentName = p.Name
		cur = p.ParentID
		depth++
	}
	resp.Depth = depth
	resp.Path = strings.Join(segments, " › ")
	return resp
}
