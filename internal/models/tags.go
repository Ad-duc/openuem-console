package models

import (
	"context"
	"strconv"

	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/app"
	"github.com/open-uem/ent/computer"
	"github.com/open-uem/ent/operatingsystem"
	"github.com/open-uem/ent/tag"
	"github.com/open-uem/ent/tenant"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (m *Model) GetAllTags(c *partials.CommonInfo, f filters.AgentFilter) ([]*ent.Tag, error) {
	var query *ent.TagQuery

	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	query = m.Client.Tag.Query().Where(tag.HasTenantWith(tenant.ID(tenantID)))

	if len(f.AgentOSVersions) > 0 {
		query.Where(tag.HasOwnerWith(agent.HasOperatingsystemWith(operatingsystem.TypeIn(f.AgentOSVersions...))))
	}

	if len(f.Nickname) > 0 {
		query.Where(tag.HasOwnerWith(agent.NicknameContainsFold(f.Nickname)))
	}

	if len(f.Username) > 0 {
		query.Where(tag.HasOwnerWith(agent.HasOperatingsystemWith(operatingsystem.UsernameContainsFold(f.Username))))
	}

	if len(f.OSVersions) > 0 {
		query.Where(tag.HasOwnerWith(agent.HasOperatingsystemWith(operatingsystem.VersionIn(f.OSVersions...))))
	}

	if len(f.ComputerManufacturers) > 0 {
		query.Where(tag.HasOwnerWith(agent.HasComputerWith(computer.ManufacturerIn(f.ComputerManufacturers...))))
	}

	if len(f.ComputerModels) > 0 {
		query.Where(tag.HasOwnerWith(agent.HasComputerWith(computer.ModelIn(f.ComputerModels...))))
	}

	if len(f.WithApplication) > 0 && len(f.WithApplicationPublisher) > 0 {
		query.Where(tag.HasOwnerWith(agent.HasComputerWith(computer.HasOwnerWith(agent.HasAppsWith(app.And(app.Name(f.WithApplication), app.Publisher(f.WithApplicationPublisher)))))))
	} else {
		if len(f.WithApplication) > 0 {
			query.Where(tag.HasOwnerWith(agent.HasComputerWith(computer.HasOwnerWith(agent.HasAppsWith(app.Name(f.WithApplication))))))
		}
		if len(f.WithApplicationPublisher) > 0 {
			query.Where(tag.HasOwnerWith(agent.HasComputerWith(computer.HasOwnerWith(agent.HasAppsWith(app.Name(f.WithApplicationPublisher))))))
		}
	}

	if len(f.IsRemote) > 0 {
		if len(f.IsRemote) == 1 && f.IsRemote[0] == "Remote" {
			query.Where(tag.HasOwnerWith(agent.HasComputerWith(computer.HasOwnerWith(agent.IsRemote(true)))))
		}

		if len(f.IsRemote) == 1 && f.IsRemote[0] == "Local" {
			query.Where(tag.HasOwnerWith(agent.HasComputerWith(computer.HasOwnerWith(agent.IsRemote(false)))))
		}
	}

	if len(f.Search) > 0 {
		query.Where(tag.HasOwnerWith(agent.Or(
			agent.NicknameContainsFold(f.Search),
			agent.OsIn(f.Search),
			agent.HasOperatingsystemWith(operatingsystem.UsernameContainsFold(f.Search)),
			agent.HasComputerWith(computer.ManufacturerContainsFold(f.Search)),
			agent.HasComputerWith(computer.ModelContainsFold(f.Search)),
		)))
	}

	return query.All(context.Background())

}

func (m *Model) GetAppliedTags(c *partials.CommonInfo) ([]*ent.Tag, error) {
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	tags, err := m.Client.Tag.Query().Where(tag.HasOwner(), tag.HasTenantWith(tenant.ID(tenantID))).All(context.Background())
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (m *Model) GetTagsByPage(p partials.PaginationAndSort, c *partials.CommonInfo) ([]*ent.Tag, error) {
	var err error
	var tags []*ent.Tag

	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	query := m.Client.Tag.Query().Where(tag.HasTenantWith(tenant.ID(tenantID))).WithOwner().Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize)

	switch p.SortBy {
	case "tag":
		if p.SortOrder == "asc" {
			query = query.Order(ent.Asc(tag.FieldTag))
		} else {
			query = query.Order(ent.Desc(tag.FieldTag))
		}
	case "description":
		if p.SortOrder == "asc" {
			query = query.Order(ent.Asc(tag.FieldDescription))
		} else {
			query = query.Order(ent.Desc(tag.FieldDescription))
		}
	default:
		query = query.Order(ent.Asc(tag.FieldID))
	}

	tags, err = query.All(context.Background())
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (m *Model) CountAllTags(c *partials.CommonInfo) (int, error) {
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return -1, err
	}

	return m.Client.Tag.Query().Where(tag.HasTenantWith(tenant.ID(tenantID))).Count(context.Background())
}

func (m *Model) NewTag(title, description, color string, c *partials.CommonInfo) error {
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return err
	}

	return m.Client.Tag.Create().SetTag(title).SetDescription(description).SetColor(color).SetTenantID(tenantID).Exec(context.Background())
}

func (m *Model) UpdateTag(tagId int, title, description, color string, c *partials.CommonInfo) error {
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return err
	}

	return m.Client.Tag.Update().SetTag(title).SetDescription(description).SetColor(color).Where(tag.ID(tagId), tag.HasTenantWith(tenant.ID(tenantID))).Exec(context.Background())
}

func (m *Model) DeleteTag(tagId int, c *partials.CommonInfo) error {
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return err
	}

	return m.Client.Tag.DeleteOneID(tagId).Where(tag.HasTenantWith(tenant.ID(tenantID))).Exec(context.Background())
}

// DefaultTag describes a tag that is created when OpenUEM is installed for the first time
type DefaultTag struct {
	Tag         string
	Description string
	Color       string
}

// DefaultTags is the list of tags created in the default organization on a fresh installation.
// Colors must be one of the names offered by the color selector of the tags view.
var DefaultTags = []DefaultTag{
	// Operating system
	{Tag: "GLOBAL - Windows", Description: "All Windows endpoints", Color: "sky"},
	{Tag: "GLOBAL - Linux", Description: "All Linux endpoints", Color: "rose"},
	{Tag: "GLOBAL - macOS", Description: "All macOS endpoints", Color: "gray"},
	// Device type
	{Tag: "GLOBAL - Servers", Description: "All servers", Color: "indigo"},
	{Tag: "GLOBAL - Desktops", Description: "All desktop computers", Color: "teal"},
	{Tag: "GLOBAL - Laptops", Description: "All laptops", Color: "cyan"},
	// Deployment rings
	{Tag: "RING - Pilot", Description: "Endpoints that receive profiles and updates first", Color: "amber"},
	{Tag: "RING - Production", Description: "Endpoints that receive profiles and updates once validated by the pilot ring", Color: "green"},
	// Lifecycle
	{Tag: "LIFECYCLE - Staging", Description: "Endpoints being prepared before they are handed over", Color: "violet"},
	{Tag: "LIFECYCLE - Decommission", Description: "Endpoints that are about to be retired", Color: "stone"},
}

// CreateDefaultTags creates the default tags for the given tenant.
// Tag names are unique for the whole database, so nothing is created if a tag
// already exists, which is the case when an existing installation is upgraded.
func (m *Model) CreateDefaultTags(tenantID int) error {
	ctx := context.Background()

	nTags, err := m.Client.Tag.Query().Count(ctx)
	if err != nil {
		return err
	}

	if nTags > 0 {
		return nil
	}

	builders := make([]*ent.TagCreate, 0, len(DefaultTags))
	for _, t := range DefaultTags {
		builders = append(builders, m.Client.Tag.Create().SetTag(t.Tag).SetDescription(t.Description).SetColor(t.Color).SetTenantID(tenantID))
	}

	return m.Client.Tag.CreateBulk(builders...).Exec(ctx)
}
