// OpenCode 管理中心 - 技能管理视图
let skills = [];
let skillsLoaded = false;

async function loadSkillsData() {
    if (skillsLoaded) return;
    skillsLoaded = true;

    try {
        skills = await api.GetSkills();
        const stats = await api.GetStats();
        renderStats(stats);
        renderSkillList();
    } catch (err) {
        skillsLoaded = false;
        showToast('加载技能数据失败: ' + (err.message || err), 'error');
    }
}

function renderStats(stats) {
    document.getElementById('statGlobal').textContent = stats.globalSkills || 0;
    api.GetSourceDir().then(function(p) {
        document.getElementById('sourcePath').textContent = p || '未知';
    }).catch(function() {
        document.getElementById('sourcePath').textContent = '未知';
    });
}

function renderSkillList(filter) {
    filter = filter || '';
    const list = document.getElementById('skillList');
    if (!skills.length) {
        list.innerHTML = '<div class="oc-empty">暂无技能</div>';
        return;
    }

    const filtered = filter
        ? skills.filter(function(s) { return s.name.toLowerCase().indexOf(filter.toLowerCase()) >= 0; })
        : skills;

    list.innerHTML = filtered.map(function(s) {
        var sourceLabel = s.source === 'project' ? '项目' : '全局';
        var sourceClass = s.source === 'project' ? 'skill-source-project' : 'skill-source-global';
        var safeName = escapeHtml(s.name);
        var safePath = escapeHtml(s.path);
        var safeDesc = escapeHtml(s.description || '无描述');
        var escapedPath = safePath.replace(/\\/g, '\\\\');
        return '<div class="skill-card" data-skill="' + safeName + '" data-path="' + safePath + '">' +
            '<div class="skill-info">' +
                '<div class="skill-name-row">' +
                    '<span class="skill-name" style="cursor:pointer;text-decoration:underline;color:var(--accent)" onclick="previewSkill(\'' + escapedPath + '\')">' + safeName + '</span>' +
                    '<span class="skill-tag ' + sourceClass + '">' + sourceLabel + '</span>' +
                '</div>' +
                '<div class="skill-desc">' + safeDesc + '</div>' +
                '<div class="skill-path">' + safePath + '</div>' +
            '</div>' +
            '<div class="skill-actions">' +
                '<button class="btn btn-sm btn-open" onclick="openSkillDir(\'' + escapedPath + '\')">📂 打开</button>' +
                '<button class="btn btn-sm" onclick="previewSkill(\'' + escapedPath + '\')">详情</button>' +
            '</div>' +
        '</div>';
    }).join('');
}

// 搜索事件在 main.js 中绑定
// 技能 Modal 事件在 DOMContentLoaded 中绑定（main.js）

// ========== 预览/编辑 Modal ==========

async function previewSkill(skillPath) {
    try {
        var result = await api.ReadSkillContent(skillPath);
        var html = marked.parse(result);
        showSkillModal('<div class="oc-text">' + html + '</div>', skillPath, false);
    } catch (err) {
        showToast('读取技能失败: ' + (err.message || err), 'error');
    }
}

async function editSkill(skillPath) {
    try {
        var result = await api.ReadSkillContent(skillPath);
        showSkillModal(escapeHtml(result), skillPath, true);
    } catch (err) {
        showToast('读取技能失败: ' + (err.message || err), 'error');
    }
}

function showSkillModal(content, skillPath, isEdit) {
    var modal = document.getElementById('skillModal');
    var body = document.getElementById('skillModalBody');
    var title = document.getElementById('skillModalTitle');
    var editBtn = document.getElementById('skillModalEdit');
    var saveBtn = document.getElementById('skillModalSave');
    var cancelBtn = document.getElementById('skillModalCancel');

    var parts = skillPath.replace(/\\/g, '/').split('/');
    title.textContent = parts[parts.length - 1] || '技能';
    modal.dataset.skillPath = skillPath;

    if (isEdit) {
        body.innerHTML = '<textarea id="skillEditArea" class="skill-edit-area">' + content + '</textarea>';
        editBtn.style.display = 'none';
        saveBtn.style.display = '';
        cancelBtn.style.display = '';
    } else {
        body.innerHTML = content;
        editBtn.style.display = '';
        saveBtn.style.display = 'none';
        cancelBtn.style.display = 'none';
    }

    modal.style.display = 'flex';
}

function closeSkillModal() {
    document.getElementById('skillModal').style.display = 'none';
}

async function saveSkillEdit() {
    var modal = document.getElementById('skillModal');
    var skillPath = modal.dataset.skillPath;
    var content = document.getElementById('skillEditArea').value;

    try {
        await api.SaveSkillContent(skillPath, content);
        showToast('保存成功', 'success');
        var html = marked.parse(content);
        showSkillModal('<div class="oc-text">' + html + '</div>', skillPath, false);
    } catch (err) {
        showToast('保存失败: ' + (err.message || err), 'error');
    }
}

// ========== Toggle 开关 ==========

async function toggleSkill(skillPath, skillName, enable) {
    try {
        var result = await api.ToggleSkill(skillPath, skillName, enable);
        if (result.success) {
            showToast((enable ? '已启用 ' : '已禁用 ') + skillName, 'success');
            skillsLoaded = false;
            loadSkillsData();
        } else {
            showToast('操作失败: ' + (result.error || '未知错误'), 'error');
        }
    } catch (err) {
        showToast('操作失败: ' + (err.message || err), 'error');
    }
}

// 打开技能目录
async function openSkillDir(skillPath) {
    try {
        await api.OpenDir(skillPath);
    } catch (err) {
        showToast('打开目录失败: ' + (err.message || err), 'error');
    }
}
