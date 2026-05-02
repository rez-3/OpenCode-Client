// ============================================================
// OpenCode 管理中心 - 技能管理视图
// ============================================================
let targets = [];
let skills = [];
let skillsLoaded = false;

async function loadSkillsData() {
    const skillList = document.getElementById('skillList');
    if (skillsLoaded) return;

    try {
        skillList.innerHTML = '<div class="loading"><div class="spinner"></div><p>正在加载技能列表...</p></div>';

        const [skillsData, targetsData, sourceDir, statsData] = await Promise.all([
            api.GetSkills(),
            api.GetTargets(),
            api.GetSourceDir(),
            api.GetStats(),
        ]);

        skills = skillsData || [];
        targets = targetsData || [];
        skillsLoaded = true;

        document.getElementById('sourcePath').textContent = (sourceDir || '未知');
        renderStats(statsData);
        renderBatchButtons();
        renderSkillList();
    } catch (err) {
        console.error('加载技能数据失败:', err);
        skillList.innerHTML = `<div class="error">
            <p>⚠️ 加载失败</p>
            <p class="error-detail">${escapeHtml(err.message || err)}</p>
            <button class="btn btn-primary" onclick="loadSkillsData()">重试</button>
        </div>`;
    }
}

function renderStats(stats) {
    if (!stats) return;
    const total = stats.totalSkills || skills.length;
    document.getElementById('statTotal').textContent = total;
}

function renderBatchButtons() {
    const batchButtons = document.getElementById('batchButtons');
    batchButtons.innerHTML = '';
    targets.forEach(target => {
        const enabledCount = skills.filter(s => s.targets[target.key]).length;
        const allEnabled = enabledCount === skills.length;

        const group = document.createElement('div');
        group.className = 'batch-group';

        const nameSpan = document.createElement('span');
        nameSpan.className = 'batch-target-name';
        nameSpan.textContent = target.label;

        const btn = document.createElement('button');
        btn.className = 'btn btn-sm ' + (allEnabled ? 'btn-danger-outline' : 'btn-success');
        btn.dataset.target = target.key;
        btn.dataset.enable = String(!allEnabled);
        btn.textContent = allEnabled ? '全部移除' : '全部启用';

        btn.addEventListener('click', async () => {
            btn.disabled = true;
            btn.textContent = '处理中...';
            try {
                const result = await api.ToggleAllSkills(target.key, !allEnabled);
                if (result.success) {
                    showToast(`${target.label} 批量操作完成`, 'success');
                    skillsLoaded = false;
                    await loadSkillsData();
                } else {
                    showToast(`部分操作失败: ${result.errors.length} 个错误`, 'error');
                }
            } catch (err) {
                showToast(`操作失败: ${err.message || err}`, 'error');
            }
        });

        group.appendChild(nameSpan);
        group.appendChild(btn);
        batchButtons.appendChild(group);
    });
}

function renderSkillList() {
    const skillList = document.getElementById('skillList');
    if (skills.length === 0) {
        skillList.innerHTML = '<div class="empty"><p>📭 没有找到技能文件</p><p class="empty-hint">请检查源目录是否有技能文件夹</p></div>';
        return;
    }

    skillList.innerHTML = '';
    skills.forEach(skill => {
        const card = createSkillCard(skill);
        skillList.appendChild(card);
    });
}

function createSkillCard(skill) {
    const card = document.createElement('div');
    card.className = 'skill-card';

    const header = document.createElement('div');
    header.className = 'skill-header';

    const info = document.createElement('div');
    info.className = 'skill-info';

    const nameEl = document.createElement('h3');
    nameEl.className = 'skill-name';
    nameEl.textContent = skill.name;

    const descEl = document.createElement('p');
    descEl.className = 'skill-desc';
    descEl.textContent = skill.description || '';

    info.appendChild(nameEl);
    if (skill.description) info.appendChild(descEl);
    header.appendChild(info);

    const toggles = document.createElement('div');
    toggles.className = 'skill-toggles';

    targets.forEach(target => {
        const isLinked = skill.targets[target.key] || false;

        const toggle = document.createElement('label');
        toggle.className = 'toggle';

        const labelSpan = document.createElement('span');
        labelSpan.className = 'toggle-label';
        labelSpan.textContent = target.label;

        const checkbox = document.createElement('input');
        checkbox.type = 'checkbox';
        checkbox.checked = isLinked;
        checkbox.dataset.skill = skill.name;
        checkbox.dataset.target = target.key;

        const slider = document.createElement('span');
        slider.className = 'toggle-slider';

        checkbox.addEventListener('change', async (e) => {
            const enable = e.target.checked;
            const skillName = e.target.dataset.skill;
            const targetKey = e.target.dataset.target;

            e.target.disabled = true;

            try {
                const result = await api.ToggleSkill(skillName, targetKey, enable);
                if (result.success) {
                    skill.targets[targetKey] = enable;
                    showToast(
                        `${skillName} → ${targets.find(t => t.key === targetKey)?.label || targetKey} ${enable ? '✅ 已启用' : '❌ 已禁用'}`,
                        'success'
                    );
                } else {
                    e.target.checked = !enable;
                    showToast(`操作失败: ${result.error || '未知错误'}`, 'error');
                }
            } catch (err) {
                e.target.checked = !enable;
                showToast(`操作失败: ${err.message || err}`, 'error');
            }

            e.target.disabled = false;
            renderBatchButtons();
        });

        toggle.appendChild(labelSpan);
        toggle.appendChild(checkbox);
        toggle.appendChild(slider);
        toggles.appendChild(toggle);
    });

    header.appendChild(toggles);
    card.appendChild(header);
    return card;
}
