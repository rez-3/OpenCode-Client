// ============================================================
// Skill Manager - 前端主逻辑
// ============================================================

// ============================================================
// 主题切换
// ============================================================
const THEME_KEY = 'skill-manager-theme';

function getTheme() {
    return localStorage.getItem(THEME_KEY) || 'light';
}

function setTheme(theme) {
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem(THEME_KEY, theme);
    updateThemeIcon(theme);
}

function toggleTheme() {
    const current = getTheme();
    setTheme(current === 'dark' ? 'light' : 'dark');
}

function updateThemeIcon(theme) {
    const btn = document.getElementById('btnTheme');
    if (btn) btn.textContent = theme === 'dark' ? '🌙' : '☀️';
}

// 启动时应用主题
setTheme(getTheme());

// ============================================================
// Wails 运行时 API 封装
const api = (() => {
    // 检查是否在 Wails 环境中运行
    if (window.go && window.go.main && window.go.main.App) {
        return window.go.main.App;
    }
// 开发模式下的 mock 数据（浏览器直接打开 HTML 时使用）
console.warn('⚠️  非 Wails 环境，使用模拟数据');
return {
    GetSkills: async () => mockSkills,
    GetTargets: async () => [
        { key: 'opencode', label: 'OpenCode', path: '~/.config/opencode/skills' },
        { key: 'claude',   label: 'Claude Code', path: '~/.claude/skills' },
        { key: 'codex',    label: 'Codex', path: '~/.codex/skills' },
    ],
    GetSourceDir: async () => '~/.cc-switch/skills',
    GetStats: async () => ({
        totalSkills: mockSkills.length,
        targetStats: { opencode: 14, claude: 6, codex: 7 },
    }),
    ToggleSkill: async (name, target, enable) => ({ skillName: name, target, linked: enable, success: true }),
    Refresh: async () => {},
    ToggleAllSkills: async (target, enable) => ({ target, enabled: enable, success: true, errors: [] }),
    // 模型配置 mock
    GetModelConfig: async () => [
        { key: 'sisyphus', type: 'agent', model: 'deepseek/deepseek-v4-pro' },
        { key: 'oracle', type: 'agent', model: 'deepseek-v4-flash' },
        { key: 'librarian', type: 'agent', model: 'deepseek-v4-flash' },
        { key: 'quick', type: 'category', model: 'deepseek-v4-flash' },
        { key: 'visual-engineering', type: 'category', model: 'deepseek-v4-flash' },
    ],
    GetAvailableModels: async () => [
        'deepseek/deepseek-chat', 'deepseek/deepseek-reasoner',
        'deepseek/deepseek-v4-flash', 'deepseek/deepseek-v4-pro',
        'openai/gpt-5.1-codex', 'openai/gpt-5.5',
    ],
    UpdateModels: async (entries) => ({ success: true }),
    GetConfigPath: async () => '~/.config/opencode/oh-my-openagent.jsonc',
    OpenDir: async (path) => { console.log('mock open:', path); },
};
})();

// 模拟数据
const mockSkills = [
    { name: 'afsim', description: 'AFSIM脚本编写助手', sourcePath: '~/.cc-switch/skills/afsim', targets: { opencode: false, claude: true, codex: true } },
    { name: 'code-review', description: '专业的代码审查助手', sourcePath: '~/.cc-switch/skills/code-review', targets: { opencode: true, claude: true, codex: true } },
    { name: 'docx', description: 'Word文档创建编辑', sourcePath: '~/.cc-switch/skills/docx', targets: { opencode: true, claude: true, codex: true } },
    { name: 'skill-creator', description: '创建新技能指南', sourcePath: '~/.cc-switch/skills/skill-creator', targets: { opencode: true, claude: true, codex: true } },
    { name: 'standards-golang', description: 'Go开发标准', sourcePath: '~/.cc-switch/skills/standards-golang', targets: { opencode: false, claude: true, codex: true } },
    { name: 'weather', description: '天气预报', sourcePath: '~/.cc-switch/skills/weather', targets: { opencode: true, claude: false, codex: false } },
    { name: 'drawio', description: '图表绘制', sourcePath: '~/.cc-switch/skills/drawio', targets: { opencode: true, claude: false, codex: false } },
];

// ============================================================
// DOM 引用
// ============================================================
const $ = (sel) => document.querySelector(sel);
const $$ = (sel) => document.querySelectorAll(sel);

const skillList = $('#skillList');
const sourcePath = $('#sourcePath');
const statsBar = $('#statsBar');
const batchButtons = $('#batchButtons');
const btnRefresh = $('#btnRefresh');
const toast = $('#toast');

// ============================================================
// Toast 提示
// ============================================================
let toastTimer = null;

function showToast(message, type = 'info') {
    toast.textContent = message;
    toast.className = `toast ${type} show`;
    clearTimeout(toastTimer);
    toastTimer = setTimeout(() => {
        toast.classList.remove('show');
    }, 2500);
}

// ============================================================
// 数据加载
// ============================================================
let targets = [];
let skills = [];

async function loadData() {
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

        sourcePath.textContent = (sourceDir || '未知');
        renderStats(statsData);
        renderBatchButtons();
        renderSkillList();
    } catch (err) {
        console.error('加载数据失败:', err);
        skillList.innerHTML = `<div class="error">
            <p>⚠️ 加载失败</p>
            <p class="error-detail">${err.message || err}</p>
            <button class="btn btn-primary" onclick="loadData()">重试</button>
        </div>`;
    }
}

// ============================================================
// 统计栏渲染
// ============================================================
function renderStats(stats) {
    if (!stats) return;

    const total = stats.totalSkills || skills.length;
    $('#statTotal').textContent = total;
}

// ============================================================
// 批量操作按钮渲染
// ============================================================
function renderBatchButtons() {
    batchButtons.innerHTML = '';
    targets.forEach(target => {
        const enabledCount = skills.filter(s => s.targets[target.key]).length;
        const allEnabled = enabledCount === skills.length;

        const group = document.createElement('div');
        group.className = 'batch-group';
        group.innerHTML = `
            <span class="batch-target-name">${target.label}</span>
            <button class="btn btn-sm ${allEnabled ? 'btn-danger-outline' : 'btn-success'}"
                    data-target="${target.key}"
                    data-enable="${!allEnabled}">
                ${allEnabled ? '全部移除' : '全部启用'}
            </button>
        `;

        const btn = group.querySelector('button');
        btn.addEventListener('click', async () => {
            btn.disabled = true;
            btn.textContent = '处理中...';
            try {
                const result = await api.ToggleAllSkills(target.key, !allEnabled);
                if (result.success) {
                    showToast(`${target.label} 批量操作完成`, 'success');
                    await loadData();
                } else {
                    showToast(`部分操作失败: ${result.errors.length} 个错误`, 'error');
                }
            } catch (err) {
                showToast(`操作失败: ${err.message || err}`, 'error');
            }
            btn.disabled = false;
        });

        batchButtons.appendChild(group);
    });
}

// ============================================================
// 技能列表渲染
// ============================================================
function renderSkillList() {
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
    info.innerHTML = `
        <h3 class="skill-name">${escapeHtml(skill.name)}</h3>
        ${skill.description ? `<p class="skill-desc">${escapeHtml(skill.description)}</p>` : ''}
    `;

    header.appendChild(info);

    const toggles = document.createElement('div');
    toggles.className = 'skill-toggles';

    targets.forEach(target => {
        const isLinked = skill.targets[target.key] || false;
        const toggle = document.createElement('label');
        toggle.className = 'toggle';
        toggle.innerHTML = `
            <span class="toggle-label">${target.label}</span>
            <input type="checkbox"
                   ${isLinked ? 'checked' : ''}
                   data-skill="${escapeHtml(skill.name)}"
                   data-target="${target.key}" />
            <span class="toggle-slider"></span>
        `;

        const checkbox = toggle.querySelector('input');
        checkbox.addEventListener('change', async (e) => {
            const enable = e.target.checked;
            const skillName = e.target.dataset.skill;
            const targetKey = e.target.dataset.target;

            // 乐观更新 UI
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
                    // 回滚 UI
                    e.target.checked = !enable;
                    showToast(`操作失败: ${(result.error || '未知错误')}`, 'error');
                }
            } catch (err) {
                // 回滚 UI
                e.target.checked = !enable;
                showToast(`操作失败: ${err.message || err}`, 'error');
            }

            e.target.disabled = false;
            // 更新批量按钮
            renderBatchButtons();
        });

        toggles.appendChild(toggle);
    });

    header.appendChild(toggles);
    card.appendChild(header);

    return card;
}

// ============================================================
// 抽屉面板
// ============================================================
const configDrawer = document.getElementById('configDrawer');
const drawerOverlay = document.getElementById('drawerOverlay');

function openDrawer() {
    configDrawer.classList.add('open');
    drawerOverlay.classList.add('open');
    loadModelConfig();
}

function closeDrawer() {
    configDrawer.classList.remove('open');
    drawerOverlay.classList.remove('open');
}

document.getElementById('btnOpenConfig').addEventListener('click', openDrawer);
document.getElementById('btnCloseDrawer').addEventListener('click', closeDrawer);
drawerOverlay.addEventListener('click', closeDrawer);

// ESC 关闭
document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape' && configDrawer.classList.contains('open')) {
        closeDrawer();
    }
});

// ============================================================
// 模型配置
// ============================================================
let modelEntries = [];
let availableModels = [];
let originalEntries = []; // 保存原始值用于变更检测

async function loadModelConfig() {
    const container = document.getElementById('modelConfig');
    container.innerHTML = '<div class="loading"><div class="spinner"></div><p>正在加载模型配置...</p></div>';

    try {
        const [entries, models, confPath] = await Promise.all([
            api.GetModelConfig(),
            api.GetAvailableModels(),
            api.GetConfigPath(),
        ]);

        modelEntries = entries.map(e => ({ ...e })); // 深拷贝可修改
        originalEntries = entries.map(e => ({ ...e })); // 保存原始值
        availableModels = models || [];

        document.getElementById('configPath').textContent = confPath || '未知';
        renderModelConfig();
    } catch (err) {
        container.innerHTML = `<div class="error">
            <p>⚠️ 加载模型配置失败</p>
            <p class="error-detail">${err.message || err}</p>
            <button class="btn btn-primary" onclick="loadModelConfig()">重试</button>
        </div>`;
    }
}

function renderModelConfig() {
    const container = document.getElementById('modelConfig');
    const actions = document.getElementById('modelActions');

    if (modelEntries.length === 0) {
        container.innerHTML = '<div class="empty"><p>📭 未找到模型配置</p></div>';
        actions.style.display = 'none';
        return;
    }

    // 分组
    const agents = modelEntries.filter(e => e.type === 'agent');
    const categories = modelEntries.filter(e => e.type === 'category');

    let html = '';

    // Agents 组
    html += '<div class="model-group"><h3 class="model-group-title">🤖 Agents</h3>';
    agents.forEach(entry => {
        html += createModelRow(entry);
    });
    html += '</div>';

    // Categories 组
    html += '<div class="model-group"><h3 class="model-group-title">📦 Categories</h3>';
    categories.forEach(entry => {
        html += createModelRow(entry);
    });
    html += '</div>';

    container.innerHTML = html;
    actions.style.display = 'flex';
    updateSaveStatus();

    // 绑定 select change 事件
    container.querySelectorAll('.model-select').forEach(select => {
        select.addEventListener('change', (e) => {
            const key = e.target.dataset.key;
            const entry = modelEntries.find(en => en.key === key);
            if (entry) {
                entry.model = e.target.value;
                updateSaveStatus();
            }
        });
    });
}

function createModelRow(entry) {
    const isChanged = originalEntries.find(o => o.key === entry.key)?.model !== entry.model;
    const options = availableModels.map(m =>
        `<option value="${m}" ${m === entry.model ? 'selected' : ''}>${m}</option>`
    ).join('');

    return `
        <div class="model-row ${isChanged ? 'changed' : ''}">
            <div class="model-row-top">
                <span class="model-key">${escapeHtml(entry.key)}</span>
                <select class="model-select" data-key="${escapeHtml(entry.key)}">${options}</select>
                <span class="model-type-badge">${entry.type}</span>
            </div>
            ${entry.comment ? `<div class="model-comment">${escapeHtml(entry.comment)}</div>` : ''}
        </div>
    `;
}

function updateSaveStatus() {
    const changedCount = modelEntries.filter(e => {
        const orig = originalEntries.find(o => o.key === e.key);
        return orig && orig.model !== e.model;
    }).length;

    const status = document.getElementById('saveStatus');
    const actions = document.getElementById('modelActions');

    if (changedCount > 0) {
        status.textContent = `${changedCount} 项未保存`;
        status.className = 'save-status changed';
    } else {
        status.textContent = '已是最新';
        status.className = 'save-status';
    }
}

// ============================================================
// 事件绑定
// ============================================================
btnRefresh.addEventListener('click', async () => {
    btnRefresh.disabled = true;
    btnRefresh.textContent = '⏳ 刷新中...';
    try {
        await api.Refresh();
        await loadData();
        showToast('列表已刷新', 'success');
    } catch (err) {
        showToast('刷新失败', 'error');
    }
    btnRefresh.disabled = false;
    btnRefresh.textContent = '🔄 刷新';
});

// 刷新模型列表按钮
document.getElementById('btnRefreshModels').addEventListener('click', async () => {
    const btn = document.getElementById('btnRefreshModels');
    btn.disabled = true;
    btn.textContent = '⏳ 刷新中...';
    try {
        availableModels = await api.GetAvailableModels();
        renderModelConfig();
        showToast(`获取到 ${availableModels.length} 个可用模型`, 'success');
    } catch (err) {
        showToast('刷新模型列表失败: ' + (err.message || err), 'error');
    }
    btn.disabled = false;
    btn.textContent = '🔄 刷新模型列表';
});

// 保存模型配置
document.getElementById('btnSaveModels').addEventListener('click', async () => {
    const btn = document.getElementById('btnSaveModels');
    const changedEntries = modelEntries.filter(e => {
        const orig = originalEntries.find(o => o.key === e.key);
        return orig && orig.model !== e.model;
    });

    if (changedEntries.length === 0) {
        showToast('没有需要保存的更改', 'info');
        return;
    }

    btn.disabled = true;
    btn.textContent = '⏳ 保存中...';

    try {
        const result = await api.UpdateModels(modelEntries);
        if (result.success) {
            originalEntries = modelEntries.map(e => ({ ...e }));
            updateSaveStatus();
            showToast(`已保存 ${changedEntries.length} 项更改`, 'success');
            // 刷新以更新变更标记
            renderModelConfig();
        } else {
            showToast('保存失败: ' + (result.error || '未知错误'), 'error');
        }
    } catch (err) {
        showToast('保存失败: ' + (err.message || err), 'error');
    }

    btn.disabled = false;
    btn.textContent = '💾 保存配置';
});

// ============================================================
// 工具函数
// ============================================================
function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// 打开目录按钮
document.getElementById('btnOpenDir').addEventListener('click', async () => {
    const path = document.getElementById('sourcePath').textContent;
    if (!path || path === '加载中...' || path === '未知') {
        showToast('目录路径无效', 'error');
        return;
    }
    try {
        await api.OpenDir(path);
    } catch (err) {
        showToast('打开目录失败: ' + (err.message || err), 'error');
    }
});

// 主题切换按钮
document.getElementById('btnTheme').addEventListener('click', toggleTheme);

// ============================================================
// 应用启动
// ============================================================
document.addEventListener('DOMContentLoaded', () => {
    loadData();
});

// Wails 运行时就绪后刷新数据
if (window.runtime) {
    window.runtime.EventsOn('wails:ready', () => {
        loadData();
    });
}
