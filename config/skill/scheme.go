package skill

import (
	"fmt"

	"oc-manager/model"
)

// ApplySkillScheme 应用技能方案：清除托管链接后逐一重建。
func (m *Manager) ApplySkillScheme(scheme model.SkillSchemeData, availableSkills []model.SkillInfo, sourceDirs []string) model.SchemeApplyResult {
	result := model.SchemeApplyResult{}

	skillMap := make(map[string]model.SkillInfo)
	for _, s := range availableSkills {
		skillMap[s.Name] = s
	}

	if err := m.ClearManagedLinks(sourceDirs); err != nil {
		// non-fatal, continue
	}

	for _, name := range scheme {
		skill, found := skillMap[name]
		if !found {
			result.Missing = append(result.Missing, name)
			continue
		}
		if skill.Conflict {
			result.Conflicts = append(result.Conflicts, name)
			continue
		}
		if err := m.LinkSkill(skill.Path, name); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", name, err.Error()))
			continue
		}
		result.Applied = append(result.Applied, name)
	}

	result.Success = len(result.Applied) > 0
	return result
}
