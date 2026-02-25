package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"

	frameEntities "github.com/leeforge/framework/entities"
)

// Post 文章实体 (examples 业务实体)
type Post struct {
	ent.Schema
}

func (Post) Mixin() []ent.Mixin {
	return []ent.Mixin{
		frameEntities.BaseEntitySchema{}, // id, created_at, updated_at, deleted_at, published_at, archived_at, owner_domain_id
	}
}

func (Post) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("author_id", uuid.UUID{}).
			Comment("文章作者ID (关联 core users 表)"),
		field.String("title").
			NotEmpty().
			Comment("文章标题"),
		field.Text("content").
			NotEmpty().
			Comment("文章内容"),
		field.String("slug").
			NotEmpty().
			Comment("URL别名"),
		field.Enum("status").
			Values("draft", "published", "archived").
			Default("draft").
			Comment("发布状态"),
		field.Int("view_count").
			Default(0).
			Comment("浏览次数"),
		field.JSON("tags", []string{}).
			Optional().
			Comment("标签"),
	}
}

func (Post) Edges() []ent.Edge {
	return nil
}

func (Post) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("owner_domain_id", "slug").Unique(),
		index.Fields("status", "published_at"),
		index.Fields("author_id"),
	}
}
