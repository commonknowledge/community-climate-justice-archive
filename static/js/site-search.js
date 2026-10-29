/**
 * Site search popup
 *
 * Opens from any `[data-el="search-open"]` button in the header. Loads
 * /filter-data.json the first time it is opened, then shows:
 * - a search input (matches story title and description)
 * - filter dropdowns (Projects, Themes, Types, What was/is/if, Weather, Gifted)
 * - 10 random suggested stories (or the matching results once the user
 *   types or picks a filter)
 */
(function () {
    const FILTERS = [
        { key: 'highStExperiment', label: 'Projects', field: 'highStExperiment', single: true },
        { key: 'themes', label: 'Themes', field: 'themes' },
        { key: 'types', label: 'Types', field: 'types' },
        { key: 'whatWasIsIf', label: 'What was/is/if', field: 'whatWasIsIf' },
        { key: 'weather', label: 'Weather', field: 'weather' },
        { key: 'giftedBy', label: 'Gifted or co-created by', field: 'giftedBy' },
    ];
    const SUGGESTION_COUNT = 10;
    const MAX_RESULTS = 30;
    const MIN_QUERY_LENGTH = 2;

    const openButtons = document.querySelectorAll('[data-el="search-open"]');
    if (openButtons.length === 0) return;

    let data = null;
    let loading = null;
    let suggestions = [];
    let modal, input, list, status, filterBar, activeBar;
    let lastFocus = null;
    const selected = {}; // filter key -> Set of selected titles

    function el(tag, className, text) {
        const node = document.createElement(tag);
        if (className) node.className = className;
        if (text !== undefined) node.textContent = text;
        return node;
    }

    function contrast(hex) {
        if (!/^#[0-9a-f]{6}$/i.test(hex || '')) return '#000';
        const r = parseInt(hex.slice(1, 3), 16);
        const g = parseInt(hex.slice(3, 5), 16);
        const b = parseInt(hex.slice(5, 7), 16);
        return (0.299 * r + 0.587 * g + 0.114 * b) / 255 > 0.5 ? '#000' : '#fff';
    }

    function storyHref(story) {
        return story.url.startsWith('/') ? story.url : '/' + story.url;
    }

    function colorFor(filterKey, title) {
        const opts = (data && data[filterKey]) || [];
        const opt = opts.find(o => o.title === title);
        return opt && opt.color ? opt.color : '#666666';
    }

    function build() {
        modal = el('div', 'site-search');
        modal.setAttribute('role', 'dialog');
        modal.setAttribute('aria-modal', 'true');
        modal.setAttribute('aria-label', 'Search the archive');
        modal.hidden = true;

        const backdrop = el('div', 'site-search__backdrop');
        backdrop.addEventListener('click', close);

        const panel = el('div', 'site-search__panel');

        const form = el('form', 'site-search__form');
        form.addEventListener('submit', e => e.preventDefault());
        input = el('input', 'site-search__input');
        input.type = 'search';
        input.placeholder = 'Search the Dudley Time Portal…';
        input.setAttribute('aria-label', 'Search the archive');
        input.autocomplete = 'off';
        input.addEventListener('input', render);
        const submit = el('button', 'site-search__submit');
        submit.type = 'submit';
        submit.setAttribute('aria-label', 'Search');
        submit.innerHTML = searchIcon();
        form.append(input, submit);

        filterBar = el('div', 'site-search__filters');
        activeBar = el('div', 'site-search__active-filters');
        activeBar.hidden = true;

        status = el('h2', 'site-search__heading caption', 'Suggestions');
        list = el('ul', 'site-search__list');

        const results = el('div', 'site-search__results');
        results.append(status, list);

        const closeBtn = el('button', 'site-search__close dashed-button', 'close');
        closeBtn.type = 'button';
        closeBtn.addEventListener('click', close);

        panel.append(closeBtn, form, filterBar, activeBar, results);
        modal.append(backdrop, panel);
        document.body.appendChild(modal);

        modal.addEventListener('keydown', e => {
            if (e.key === 'Escape') close();
        });
        panel.addEventListener('click', e => {
            if (!e.target.closest('.filter-dropdown')) closeDropdowns();
        });
    }

    function searchIcon() {
        return '<svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M15.5 14h-.79l-.28-.27A6.47 6.47 0 0 0 16 9.5 6.5 6.5 0 1 0 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"/></svg>';
    }

    function closeDropdowns() {
        modal.querySelectorAll('.filter-dropdown.open').forEach(d => {
            d.classList.remove('open');
            d.querySelector('button').setAttribute('aria-expanded', 'false');
        });
    }

    function buildFilters() {
        filterBar.innerHTML = '';
        FILTERS.forEach(f => {
            const options = data[f.key] || [];
            if (options.length === 0) return;
            selected[f.key] = selected[f.key] || new Set();

            const wrap = el('div', 'filter-dropdown filter-dropdown--small filter-dropdown--white');
            const trigger = el('button', 'filter-dropdown__trigger dashed-button', f.label);
            trigger.type = 'button';
            trigger.setAttribute('aria-expanded', 'false');
            trigger.dataset.label = f.label;
            const menu = el('div', 'filter-dropdown__options');

            [...options]
                .sort((a, b) => a.title.localeCompare(b.title))
                .forEach(opt => {
                    const b = el('button', 'filter-dropdown__option', opt.title);
                    b.type = 'button';
                    b.dataset.filter = f.key;
                    b.dataset.value = opt.title;
                    const color = opt.color || '#666666';
                    b.style.backgroundColor = color;
                    b.style.color = contrast(color);
                    b.addEventListener('click', e => {
                        e.stopPropagation();
                        const set = selected[f.key];
                        if (set.has(opt.title)) set.delete(opt.title);
                        else set.add(opt.title);
                        closeDropdowns();
                        renderActiveFilters();
                        render();
                    });
                    menu.appendChild(b);
                });

            trigger.addEventListener('click', e => {
                e.stopPropagation();
                const wasOpen = wrap.classList.contains('open');
                closeDropdowns();
                if (!wasOpen) {
                    wrap.classList.add('open');
                    trigger.setAttribute('aria-expanded', 'true');
                }
            });

            wrap.append(trigger, menu);
            filterBar.appendChild(wrap);
        });
    }

    // Selected filters shown as removable tags below the dropdowns
    function renderActiveFilters() {
        activeBar.innerHTML = '';
        FILTERS.forEach(f => {
            (selected[f.key] ? [...selected[f.key]] : []).forEach(title => {
                const color = colorFor(f.key, title);
                const tag = el('button', 'site-search__active-tag tag', title);
                tag.type = 'button';
                tag.style.backgroundColor = color;
                tag.style.color = contrast(color);
                const x = el('span', 'tag__remove', '×');
                x.setAttribute('aria-label', 'Remove filter');
                tag.appendChild(x);
                tag.addEventListener('click', () => {
                    selected[f.key].delete(title);
                    renderActiveFilters();
                    render();
                });
                activeBar.appendChild(tag);
            });
        });
        activeBar.hidden = activeBar.children.length === 0;

        // Mark selected options in the dropdowns
        filterBar.querySelectorAll('.filter-dropdown__option').forEach(o => {
            o.classList.toggle('selected', selected[o.dataset.filter].has(o.dataset.value));
        });
    }

    function hasFilters() {
        return Object.values(selected).some(s => s.size > 0);
    }

    function matches(story, terms) {
        for (const f of FILTERS) {
            const set = selected[f.key];
            if (!set || set.size === 0) continue;
            const value = story[f.field];
            const values = Array.isArray(value) ? value : (value ? [value] : []);
            if (!values.some(v => set.has(v))) return false;
        }
        if (terms.length === 0) return true;
        const haystack = [story.finding, story.experience].join(' ').toLowerCase();
        return terms.every(t => haystack.includes(t));
    }

    function randomSample(items, count) {
        const pool = [...items];
        for (let i = pool.length - 1; i > 0; i--) {
            const j = Math.floor(Math.random() * (i + 1));
            [pool[i], pool[j]] = [pool[j], pool[i]];
        }
        return pool.slice(0, count);
    }

    function pickSuggestions() {
        suggestions = randomSample(data.stories, SUGGESTION_COUNT);
    }

    // Appends text to parent, wrapping any search term matches in <mark>
    function appendHighlighted(parent, text, terms) {
        if (terms.length === 0) {
            parent.textContent = text;
            return;
        }
        const escaped = terms.map(t => t.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'));
        const re = new RegExp('(' + escaped.join('|') + ')', 'gi');
        text.split(re).forEach((part, i) => {
            if (!part) return;
            if (i % 2 === 1) parent.appendChild(el('mark', 'site-search__mark', part));
            else parent.appendChild(document.createTextNode(part));
        });
    }

    function render() {
        if (!data) return;
        const query = input.value.trim().toLowerCase();
        // Ignore text searches until at least 2 characters are typed
        const terms = query.length >= MIN_QUERY_LENGTH ? query.split(/\s+/).filter(Boolean) : [];
        const searching = terms.length > 0 || hasFilters();
        let stories;
        if (searching) {
            const all = data.stories.filter(s => matches(s, terms));
            if (terms.length > 0) {
                const inTitle = s => {
                    const title = (s.finding || '').toLowerCase();
                    return terms.every(t => title.includes(t));
                };
                // Array.sort is stable, so the original order is kept within each group
                all.sort((a, b) => inTitle(b) - inTitle(a));
            }
            stories = all.slice(0, MAX_RESULTS);
            status.textContent = all.length === 0
                ? 'No results'
                : all.length + (all.length === 1 ? ' result' : ' results');
        } else {
            stories = suggestions;
            status.textContent = 'Suggestions';
        }

        list.innerHTML = '';
        stories.forEach(story => {
            const li = el('li', 'site-search__item');
            const link = el('a', 'site-search__link');
            appendHighlighted(link, story.finding || 'Untitled', terms);
            link.href = storyHref(story);
            const tags = el('span', 'site-search__tags');
            story._tags = story._tags || randomSample(story.themes || [], 3);
            story._tags.forEach(t => {
                const tag = el('span', 'tag', t);
                const color = colorFor('themes', t);
                tag.style.backgroundColor = color;
                tag.style.color = contrast(color);
                tags.appendChild(tag);
            });
            li.append(link, tags);
            list.appendChild(li);
        });
    }

    function load() {
        if (data) return Promise.resolve();
        if (!loading) {
            loading = fetch('/filter-data.json')
                .then(r => {
                    if (!r.ok) throw new Error('Failed to load search data: ' + r.status);
                    return r.json();
                })
                .then(json => {
                    data = json;
                    data.stories = data.stories || [];
                    pickSuggestions();
                    buildFilters();
                })
                .catch(err => {
                    loading = null;
                    console.error(err);
                    status.textContent = 'Search is unavailable right now.';
                });
        }
        return loading;
    }

    function open() {
        if (!modal) build();
        lastFocus = document.activeElement;
        modal.hidden = false;
        document.body.style.overflow = 'hidden';
        input.focus();
        load().then(render);
    }

    function close() {
        if (!modal || modal.hidden) return;
        closeDropdowns();
        modal.hidden = true;
        document.body.style.overflow = '';
        if (lastFocus && lastFocus.focus) lastFocus.focus();
    }

    openButtons.forEach(btn => {
        btn.addEventListener('click', e => {
            e.preventDefault();
            // Close the mobile menu overlay if the button was inside it
            const overlay = btn.closest('.mobile-menu-overlay');
            if (overlay) {
                overlay.classList.remove('is-open');
                overlay.setAttribute('aria-hidden', 'true');
            }
            open();
        });
    });
})();
