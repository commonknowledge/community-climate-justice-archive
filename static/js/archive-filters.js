/**
 * Archive Filtering System
 * 
 * This JavaScript makes the filtering work on the Archive page. When you click
 * filter buttons (like "Climate Change" or "Photo"), it shows only the stories
 * that match your selections.
 * 
 * How it works:
 * 1. When the page loads, fetch a JSON file with all the story data
 * 2. Display all stories initially
 * 3. When someone clicks a filter, hide stories that don't match
 * 4. Update the URL so you can bookmark or share filtered views
 * 5. Lazy-load more stories as you scroll down (performance optimization)
 * 
 * This all happens in the browser ("client-side") using vanilla JavaScript -
 * no libraries like jQuery, no server requests after the initial load. This
 * makes filtering instant and reduces server load.
 * 
 * The code is organized as a class called ArchiveFilters. When the page loads,
 * we create a single instance of this class which handles everything.
 */

class ArchiveFilters {
    /**
     * Set up the filtering system
     * 
     * This constructor runs once when the page loads. It:
     * - Finds all the HTML elements we need (buttons, dropdowns, etc.)
     * - Loads the story data from a JSON file
     * - Sets up event listeners for user interactions
     * - Checks if the URL has filter parameters and applies them
     */
    constructor() {
        this.filterData = null;
        this.currentFilters = {
            themes: [],
            types: [],
            weather: [],
            whatWasIsIf: [],
            scalePermanence: [],
            timePeriod: []
        };
        this.filteredStories = [];
        this.currentSort = 'dateExperienced';
        this.currentPage = 0;
        this.storiesPerPage = 20;

        this.initializeElements();
        this.loadFilterData();
        this.setupEventListeners();
        this.loadFiltersFromURL();
    }
    
    /**
     * Find all the HTML elements we need
     * 
     * We store references to buttons, dropdowns, and containers so we can
     * manipulate them later without searching the page every time.
     * 
     * The code checks if each element exists and warns us if something is missing.
     * This helps catch problems during development.
     */
    initializeElements() {
        // Store all the HTML elements we'll need
        const el = (name) => document.querySelector(`[data-el="${name}"]`);
        this.elements = {
            // Primary filter dropdowns
            themeDropdown: el('theme-dropdown'),
            typeDropdown: el('type-dropdown'),
            whatwasisifDropdown: el('whatwasisif-dropdown'),
            weatherDropdown: el('weather-dropdown'),
            // More filter dropdowns
            scalepermanenceDropdown: el('scalepermanence-dropdown'),
            timeperiodDropdown: el('timeperiod-dropdown'),
            // Buttons that open the dropdowns
            themeButton: el('theme-button'),
            typeButton: el('type-button'),
            whatwasisifButton: el('whatwasisif-button'),
            weatherButton: el('weather-button'),
            scalepermanenceButton: el('scalepermanence-button'),
            timeperiodButton: el('timeperiod-button'),
            // Content areas inside the dropdowns
            themeContent: el('theme-content'),
            typeContent: el('type-content'),
            whatwasisifContent: el('whatwasisif-content'),
            weatherContent: el('weather-content'),
            scalepermanenceContent: el('scalepermanence-content'),
            timeperiodContent: el('timeperiod-content'),
            // More filters toggle
            moreFiltersToggle: el('more-filters-toggle'),
            moreFiltersContainer: el('more-filters'),
            // Sort elements
            sortDropdown: el('sort-dropdown'),
            sortButton: el('sort-button'),
            sortContent: el('sort-content'),
            sortText: el('sort-text'),
            // The "Clear filters" button
            clearFilters: el('clear-filters'),
            // Text showing "X of Y stories"
            filterCount: el('filter-count'),
            // The area showing active filter tags
            activeFilters: el('active-filters'),
            activeFiltersList: el('active-filters-list'),
            // Where the story grid gets displayed
            storiesContainer: el('stories-container'),
            // The total count in the header
            totalCount: el('total-count'),
            // The archive heading (text changes when filters are active)
            archiveHeading: el('archive-heading')
        };
        
        // Safety check - warn if any elements are missing
        for (const [name, element] of Object.entries(this.elements)) {
            if (!element) {
                console.warn(`Filter element not found: ${name}`);
            }
        }
    }
    
    /**
     * Load story data from the JSON file
     * 
     * When the archive is built, the Go program creates a file called
     * filter-data.json containing all the stories and their tags. This
     * function fetches that file and stores the data.
     * 
     * Once we have the data, we:
     * - Fill in the filter dropdowns with all available options
     * - Set up the initial list of stories to display
     * - Update the count showing how many stories there are
     */
    async loadFilterData() {
        try {
            // Fetch the JSON file (this is the only network request after page load)
            const response = await fetch('/filter-data.json');
            if (!response.ok) {
                throw new Error(`Failed to load filter data: ${response.status}`);
            }
            
            // Parse the JSON and store it
            this.filterData = await response.json();
            
            // Fill in the dropdown menus with filter options
            this.populateFilterDropdowns();
            
            // Start with all stories visible
            this.filteredStories = [...this.filterData.stories];

            // Re-apply URL filters (if any), then render stories/count from JSON data
            this.updateDropdownDisplay();
            this.applyFilters();
            this.updateActiveFiltersDisplay();
            
        } catch (error) {
            console.error('Error loading filter data:', error);
            // Show a friendly error message to the user
            this.showError('Failed to load filtering options. Please refresh the page.');
        }
    }
    
    populateFilterDropdowns() {
        if (!this.filterData) return;

        // Populate primary filters
        this.populateDropdown(this.elements.themeContent, this.filterData.themes, 'themes');
        this.populateDropdown(this.elements.typeContent, this.filterData.types, 'types');
        this.populateDropdown(this.elements.whatwasisifContent, this.filterData.whatWasIsIf, 'whatWasIsIf');
        this.populateDropdown(this.elements.weatherContent, this.filterData.weather, 'weather');

        // Populate "more" filters
        this.populateDropdown(this.elements.scalepermanenceContent, this.filterData.scalePermanence, 'scalePermanence');
        this.populateDropdown(this.elements.timeperiodContent, this.filterData.timePeriod, 'timePeriod');

        // Show "More filters" toggle only if any of the extra filter types have data
        const hasMoreFilters = (this.filterData.scalePermanence && this.filterData.scalePermanence.length > 0) ||
            (this.filterData.timePeriod && this.filterData.timePeriod.length > 0);

        if (hasMoreFilters && this.elements.moreFiltersToggle) {
            this.elements.moreFiltersToggle.style.display = '';
        }
    }
    
    populateDropdown(contentElement, options, filterType) {
        if (!contentElement || !options) return;

        // Clear existing content
        contentElement.innerHTML = '';

        if (options.length === 0) return;

        // Add options sorted by title
        const sortedOptions = [...options].sort((a, b) => a.title.localeCompare(b.title));

        sortedOptions.forEach(option => {
            const tagButton = document.createElement('button');
            tagButton.className = 'filter-dropdown__option';

            // Get color - use default if empty
            let color = option.color;
            if (!color || color === '') {
                color = this.getDefaultColor(option.title, filterType);
            }

            tagButton.style.backgroundColor = color;
            tagButton.style.color = this.getContrastColor(color);
            tagButton.textContent = option.title;
            tagButton.dataset.value = option.title;
            tagButton.dataset.filterType = filterType;

            tagButton.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();
                this.toggleFilter(option.title, filterType);
            });

            contentElement.appendChild(tagButton);
        });
    }
    
    // Helper function to get default colors when none are provided
    getDefaultColor(title, filterType) {
        if (filterType === 'weather') {
            // Assign sensible colors for weather conditions
            const weatherColors = {
                'sunny': '#FFD700',       // Gold
                'cloudy': '#87CEEB',      // Sky blue
                'rainy': '#4682B4',       // Steel blue
                'overcast': '#708090',    // Slate gray
                'pouring': '#191970',     // Midnight blue
                'drizzly': '#6495ED',     // Cornflower blue
                'wet': '#4169E1',         // Royal blue
                'damp': '#5F9EA0',        // Cadet blue
                'grey': '#A9A9A9',        // Dark gray
                'foggy': '#D3D3D3',       // Light gray
                'dark': '#2F4F4F',        // Dark slate gray
                'cold': '#B0E0E6',        // Powder blue
                'warm': '#FFA500',        // Orange
                'hot': '#FF6347',         // Tomato
                'mild': '#98FB98',        // Pale green
                'cool': '#ADD8E6',        // Light blue
                'dry': '#F5DEB3',         // Wheat
                'chilled': '#E0FFFF',     // Light cyan
                'evening': '#483D8B',     // Dark slate blue
                'dank': '#556B2F',        // Dark olive green
                'fairy rain': '#E6E6FA',  // Lavender
                'slight breeze': '#F0F8FF', // Alice blue
                'might rain later': '#9370DB' // Medium purple
            };
            
            // Try exact match first
            const lowerTitle = title.toLowerCase();
            if (weatherColors[lowerTitle]) {
                return weatherColors[lowerTitle];
            }
            
            // Try partial matches for compound weather descriptions
            for (const [key, color] of Object.entries(weatherColors)) {
                if (lowerTitle.includes(key)) {
                    return color;
                }
            }
            
            // Default weather color
            return '#87CEEB'; // Sky blue
        }
        
        // Default color for other filter types
        return '#666666'; // Dark gray
    }
    
    // Helper function to determine if text should be black or white based on background
    getContrastColor(hexColor) {
        // Convert hex to RGB
        const r = parseInt(hexColor.slice(1, 3), 16);
        const g = parseInt(hexColor.slice(3, 5), 16);
        const b = parseInt(hexColor.slice(5, 7), 16);
        
        // Calculate luminance
        const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255;
        
        return luminance > 0.5 ? '#000000' : '#ffffff';
    }
    
    // Helper function to get tag color from filter data
    getTagColor(tagTitle, filterType) {
        if (!this.filterData || !this.filterData[filterType]) {
            return this.getDefaultColor(tagTitle, filterType);
        }
        
        const option = this.filterData[filterType].find(opt => opt.title === tagTitle);
        if (option && option.color) {
            return option.color;
        }
        
        return this.getDefaultColor(tagTitle, filterType);
    }
    
    setupEventListeners() {
        // Dropdown button handlers
        if (this.elements.themeButton) {
            this.elements.themeButton.addEventListener('click', (e) => {
                e.preventDefault();
                this.toggleDropdown('theme');
            });
        }
        
        if (this.elements.typeButton) {
            this.elements.typeButton.addEventListener('click', (e) => {
                e.preventDefault();
                this.toggleDropdown('type');
            });
        }
        
        if (this.elements.weatherButton) {
            this.elements.weatherButton.addEventListener('click', (e) => {
                e.preventDefault();
                this.toggleDropdown('weather');
            });
        }

        if (this.elements.whatwasisifButton) {
            this.elements.whatwasisifButton.addEventListener('click', (e) => {
                e.preventDefault();
                this.toggleDropdown('whatwasisif');
            });
        }

        if (this.elements.scalepermanenceButton) {
            this.elements.scalepermanenceButton.addEventListener('click', (e) => {
                e.preventDefault();
                this.toggleDropdown('scalepermanence');
            });
        }

        if (this.elements.timeperiodButton) {
            this.elements.timeperiodButton.addEventListener('click', (e) => {
                e.preventDefault();
                this.toggleDropdown('timeperiod');
            });
        }

        // More filters toggle
        if (this.elements.moreFiltersToggle) {
            this.elements.moreFiltersToggle.addEventListener('click', () => {
                const container = this.elements.moreFiltersContainer;
                if (!container) return;
                const isOpen = container.style.display !== 'none';
                container.style.display = isOpen ? 'none' : '';
                this.elements.moreFiltersToggle.classList.toggle('open', !isOpen);
                const textEl = this.elements.moreFiltersToggle.querySelector('.more-filters-text');
                if (textEl) {
                    textEl.textContent = isOpen ? 'More filters' : 'Fewer filters';
                }
            });
        }

        // Sort dropdown toggle
        if (this.elements.sortButton) {
            this.elements.sortButton.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();
                const dropdown = this.elements.sortDropdown;
                if (!dropdown) return;
                const isOpen = dropdown.classList.contains('open');
                this.closeAllDropdowns();
                this.closeSortDropdown();
                if (!isOpen) {
                    dropdown.classList.add('open');
                }
            });
        }

        // Sort option selection
        if (this.elements.sortContent) {
            this.elements.sortContent.addEventListener('click', (e) => {
                const option = e.target.closest('.sort-option');
                if (!option) return;
                e.stopPropagation();
                this.setSort(option.dataset.sort);
            });
        }

        // Close dropdowns when clicking outside
        document.addEventListener('click', (e) => {
            if (!e.target.closest('.filter-dropdown') && !e.target.closest('.sort-dropdown')) {
                this.closeAllDropdowns();
                this.closeSortDropdown();
            }
        });
        
        // Clear filters button
        if (this.elements.clearFilters) {
            this.elements.clearFilters.addEventListener('click', () => this.clearAllFilters());
        }
        
        // Handle browser back/forward
        window.addEventListener('popstate', () => this.loadFiltersFromURL());
        
        // Handle scroll for lazy loading
        window.addEventListener('scroll', this.debounce(() => this.handleScroll(), 100));
    }
    
    toggleDropdown(dropdownType) {
        const dropdown = this.elements[`${dropdownType}Dropdown`];
        if (!dropdown) return;

        const isOpen = dropdown.classList.contains('open');

        // Close all dropdowns (including sort) first
        this.closeAllDropdowns();

        // Toggle the clicked dropdown
        if (!isOpen) {
            dropdown.classList.add('open');
        }
    }
    
    closeAllDropdowns() {
        ['theme', 'type', 'weather', 'whatwasisif', 'scalepermanence', 'timeperiod'].forEach(type => {
            const dropdown = this.elements[`${type}Dropdown`];
            if (dropdown) {
                dropdown.classList.remove('open');
            }
        });
        this.closeSortDropdown();
    }

    closeSortDropdown() {
        if (this.elements.sortDropdown) {
            this.elements.sortDropdown.classList.remove('open');
        }
    }

    setSort(sortKey) {
        this.currentSort = sortKey;

        // Update button text
        const labels = {
            dateExperienced: 'Date experienced',
            dateCreated: 'Date created',
            random: 'Random'
        };
        if (this.elements.sortText) {
            this.elements.sortText.textContent = labels[sortKey] || sortKey;
        }

        // Update selected state in dropdown
        if (this.elements.sortContent) {
            this.elements.sortContent.querySelectorAll('.sort-option').forEach(opt => {
                opt.classList.toggle('selected', opt.dataset.sort === sortKey);
            });
        }

        this.closeSortDropdown();
        this.sortStories();
        this.currentPage = 0;
        this.renderStories();
        this.updateURL();
    }

    sortStories() {
        if (this.currentSort === 'random') {
            // Fisher-Yates shuffle
            for (let i = this.filteredStories.length - 1; i > 0; i--) {
                const j = Math.floor(Math.random() * (i + 1));
                [this.filteredStories[i], this.filteredStories[j]] = [this.filteredStories[j], this.filteredStories[i]];
            }
        } else if (this.currentSort === 'dateExperienced') {
            this.filteredStories.sort((a, b) => {
                const dateA = a.startDateTime || '';
                const dateB = b.startDateTime || '';
                return dateB.localeCompare(dateA);
            });
        } else if (this.currentSort === 'dateCreated') {
            this.filteredStories.sort((a, b) => {
                const dateA = a.createdTime || '';
                const dateB = b.createdTime || '';
                return dateB.localeCompare(dateA);
            });
        }
    }

    toggleFilter(value, filterType) {
        const filters = this.currentFilters[filterType];
        const index = filters.indexOf(value);
        
        if (index > -1) {
            // Remove filter
            filters.splice(index, 1);
        } else {
            // Add filter
            filters.push(value);
        }
        
        this.updateDropdownDisplay();
        this.applyFilters();
        this.updateURL();
        this.updateActiveFiltersDisplay();
        
        // Close the dropdown after selection
        this.closeAllDropdowns();
    }
    
    updateDropdownDisplay() {
        // Update the visual state of filter options to show which are selected
        const filterTypeToElement = {
            themes: this.elements.themeContent,
            types: this.elements.typeContent,
            weather: this.elements.weatherContent,
            whatWasIsIf: this.elements.whatwasisifContent,
            scalePermanence: this.elements.scalepermanenceContent,
            timePeriod: this.elements.timeperiodContent
        };
        ['themes', 'types', 'weather', 'whatWasIsIf', 'scalePermanence', 'timePeriod'].forEach(filterType => {
            const contentElement = filterTypeToElement[filterType];
            if (!contentElement) return;
            
            const options = contentElement.querySelectorAll('.filter-dropdown__option');
            options.forEach(option => {
                const isSelected = this.currentFilters[filterType].includes(option.dataset.value);
                option.classList.toggle('selected', isSelected);
            });
        });
    }
    
    /**
     * Filter the stories based on current selections
     * 
     * This is the heart of the filtering system. It goes through all the stories
     * and checks if each one matches the selected filters.
     * 
     * For a story to be shown, it needs to match:
     * - At least one of the selected themes (if any themes are selected)
     * - At least one of the selected types (if any types are selected)
     * - At least one of the selected weather conditions (if any are selected)
     * 
     * So if you select "Climate Change" and "Photo", you'll see all photos about
     * climate change, even if they also have other themes or types.
     * 
     * After filtering, we re-render the grid to show only matching stories.
     */
    applyFilters() {
        if (!this.filterData) return;
        
        // Go through all stories and keep only the ones that match
        this.filteredStories = this.filterData.stories.filter(story => {
            // Check theme filters
            if (this.currentFilters.themes.length > 0) {
                // Does this story have at least one of the selected themes?
                const hasMatchingTheme = this.currentFilters.themes.some(theme => 
                    story.themes.includes(theme)
                );
                if (!hasMatchingTheme) return false;  // Doesn't match, hide it
            }
            
            // Check type filters
            if (this.currentFilters.types.length > 0) {
                // Does this story have at least one of the selected types?
                const hasMatchingType = this.currentFilters.types.some(type => 
                    story.types.includes(type)
                );
                if (!hasMatchingType) return false;  // Doesn't match, hide it
            }
            
            // Check weather filters
            if (this.currentFilters.weather.length > 0) {
                // Does this story have at least one of the selected weather conditions?
                const hasMatchingWeather = this.currentFilters.weather.some(weather =>
                    story.weather.includes(weather)
                );
                if (!hasMatchingWeather) return false;  // Doesn't match, hide it
            }

            // Check what was/is/if filters
            if (this.currentFilters.whatWasIsIf.length > 0) {
                const hasMatchingWhatWasIsIf = this.currentFilters.whatWasIsIf.some(wwii =>
                    story.whatWasIsIf.includes(wwii)
                );
                if (!hasMatchingWhatWasIsIf) return false;
            }

            // Check scale permanence filters
            if (this.currentFilters.scalePermanence.length > 0) {
                const hasMatchingScalePermanence = this.currentFilters.scalePermanence.some(sp =>
                    story.scalePermanence.includes(sp)
                );
                if (!hasMatchingScalePermanence) return false;
            }

            // Check time period filters
            if (this.currentFilters.timePeriod.length > 0) {
                const hasMatchingTimePeriod = this.currentFilters.timePeriod.some(tp =>
                    story.timePeriod.includes(tp)
                );
                if (!hasMatchingTimePeriod) return false;
            }

            // If we get here, the story matches all filters
            return true;
        });
        
        // Apply current sort
        this.sortStories();

        // Reset to the first page
        this.currentPage = 0;

        // Redraw the story grid with the filtered results
        this.renderStories();
        
        // Update the "X of Y stories" text
        this.updateFilterCount();
    }
    
    /**
     * Display the filtered stories on the page
     * 
     * This function rebuilds the story grid with the current filtered stories.
     * It doesn't show all stories at once - instead it uses "lazy loading" to
     * show 20 stories at a time, loading more as you scroll down. This keeps
     * the page fast even with hundreds of stories.
     * 
     * We add a short delay (50ms) to show a loading state, which gives visual
     * feedback that something is happening when you change filters.
     */
    renderStories() {
        if (!this.elements.storiesContainer) return;
        
        // Add a CSS class to show we're filtering (CSS can show a loading spinner)
        this.elements.storiesContainer.classList.add('filtering');
        
        // Brief delay to let the loading state show (better user experience)
        setTimeout(() => {
            // Clean up any popups that were created by previous filtering
            const dynamicPopups = document.querySelectorAll('.story-popup[data-dynamic="true"]');
            dynamicPopups.forEach(popup => popup.remove());
            
            // Remove all the current stories from the page
            this.elements.storiesContainer.innerHTML = '';
            
            if (this.filteredStories.length === 0) {
                // No matches - show a friendly message
                this.elements.storiesContainer.appendChild(this.createNoResultsElement());
            } else {
                // Show the first batch of stories (more load as you scroll)
                this.renderStoriesBatch(0, this.storiesPerPage);
            }
            
            // Remove the loading state
            this.elements.storiesContainer.classList.remove('filtering');
        }, 50);
    }
    
    renderStoriesBatch(startIndex, count) {
        const endIndex = Math.min(startIndex + count, this.filteredStories.length);
        const fragment = document.createDocumentFragment();
        
        for (let i = startIndex; i < endIndex; i++) {
            const story = this.filteredStories[i];
            const storyElement = this.createStoryElement(story);
            fragment.appendChild(storyElement);
        }
        
        this.elements.storiesContainer.appendChild(fragment);
    }
    
    createStoryElement(story) {
        const storyDiv = document.createElement('div');
        storyDiv.className = 'story';
        
        if (story.attachment && story.attachment.url) {
            const attachment = story.attachment;
            let content = '';
            
            if (attachment.fileType === 'image') {
                content = `
                    <a class="story-image-container" href="${story.url}" data-story-id="${story.id}">
                        <img 
                            src="${attachment.thumbUrl}" 
                            srcset="${attachment.thumbUrl} 300w, ${attachment.mediumUrl} 800w"
                            sizes="(max-width: 600px) 300px, 800px"
                            alt="${attachment.alt || ''}"
                            loading="lazy">
                    </a>
                `;
            } else if (attachment.fileType === 'video') {
                content = `
                    <a class="story-image-container" href="${story.url}" data-story-id="${story.id}">
                        <div class="story-video-preview">
                            <video class="story-video-preview__thumb" muted playsinline preload="metadata" aria-hidden="true">
                                <source src="${attachment.url}#t=0.1" type="${attachment.type || 'video/mp4'}">
                            </video>
                            <div class="story-video-preview__overlay" aria-hidden="true">
                                <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                    <rect x="2" y="4" width="20" height="16" rx="2"></rect>
                                    <polygon points="10,9 16,12 10,15 10,9"></polygon>
                                </svg>
                            </div>
                        </div>
                    </a>
                `;
            } else if (attachment.fileType === 'audio') {
                content = `
                    <a class="story-image-container" href="${story.url}" data-story-id="${story.id}">
                        <div class="story-audio-preview">
                            <div class="audio-play-button">
                                <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                    <circle cx="12" cy="12" r="10"></circle>
                                    <polygon points="10,8 16,12 10,16 10,8"></polygon>
                                </svg>
                            </div>
                            <div class="audio-filename">${attachment.filename}</div>
                        </div>
                    </a>
                `;
            } else if (attachment.fileType === 'document') {
                content = `
                    <a class="story-image-container" href="${story.url}" data-story-id="${story.id}">
                        <div class="story-document-preview">
                            <div class="document-icon">
                                <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
                                    <polyline points="14,2 14,8 20,8"></polyline>
                                    <line x1="16" y1="13" x2="8" y2="13"></line>
                                    <line x1="16" y1="17" x2="8" y2="17"></line>
                                    <polyline points="10,9 9,9 8,9"></polyline>
                                </svg>
                            </div>
                            <div class="document-filename">${attachment.filename}</div>
                        </div>
                    </a>
                `;
            }
            
            storyDiv.innerHTML = content;
        } else {
            // Text-only story - show the finding text in the card
            storyDiv.innerHTML = `
                <a class="story-image-container story-text-only-card" href="${story.url}" data-story-id="${story.id}">
                    <div class="story-text-only-preview">
                        <p class="story-text-only-preview-text">${story.finding}</p>
                    </div>
                </a>
            `;
        }
        
        // Create the corresponding popup element
        const popupDiv = document.createElement('div');
        popupDiv.className = 'story-popup';
        popupDiv.setAttribute('data-story-popup-id', story.id);
        popupDiv.setAttribute('data-dynamic', 'true');
        
        let popupContent = '';
        if (story.attachment && story.attachment.url) {
            const attachment = story.attachment;
            if (attachment.fileType === 'image') {
                popupContent = `<img data-src="${attachment.largeUrl || attachment.url}" alt="" class="popup-img">`;
            } else if (attachment.fileType === 'video') {
                popupContent = `
                    <div></div>
                `;
            } else if (attachment.fileType === 'audio') {
                popupContent = `
                    <div class="popup-audio">
                        <audio controls>
                            <source src="${attachment.url}" type="${attachment.type || 'audio/mpeg'}">
                            Your browser does not support the audio element.
                        </audio>
                    </div>
                `;
            } else if (attachment.fileType === 'document') {
                popupContent = `
                    <div class="popup-document">
                        <div class="document-icon-large">
                            <svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
                                <polyline points="14,2 14,8 20,8"></polyline>
                                <line x1="16" y1="13" x2="8" y2="13"></line>
                                <line x1="16" y1="17" x2="8" y2="17"></line>
                                <polyline points="10,9 9,9 8,9"></polyline>
                            </svg>
                        </div>
                        <a href="${attachment.url}" download="${attachment.filename}" class="download-link">
                            Download ${attachment.filename}
                        </a>
                    </div>
                `;
            }
        } else {
            // Text-only story - show text in popup
            popupContent = `
                <div class="popup-text-only">
                    <p class="popup-text-only-content">${story.finding}</p>
                </div>
            `;
        }
        
        // Build tag content with proper colors (sorted alphabetically)
        let tagContent = '';

        const sortedTypes = this.sortAlphabetically(story.types);
        for (const type of sortedTypes) {
            const color = this.getTagColor(type, 'types');
            tagContent += `<span class="tag" style="background-color: ${color};">${type}</span>`;
        }

        const sortedWeather = this.sortAlphabetically(story.weather);
        for (const weather of sortedWeather) {
            const color = this.getTagColor(weather, 'weather');
            tagContent += `<span class="tag" style="background-color: ${color};">${weather}</span>`;
        }

        const sortedThemes = this.sortAlphabetically(story.themes);
        for (const theme of sortedThemes) {
            const color = this.getTagColor(theme, 'themes');
            tagContent += `<span class="tag" style="background-color: ${color};">${theme}</span>`;
        }
        
        popupDiv.innerHTML = `
            <div class="story-popup-container">
                ${popupContent}
                <div class="story-popup-text">
                    <div class="story-popup-text-finding">
                        <span class="story-popup-text-finding-content">
                            ${story.finding}
                        </span>
                    </div>
                    <div class="story-popup-text-tags">
                        ${tagContent}
                    </div>
                </div>
            </div>
        `;
        
        // Append popup to body (where other popups are)
        document.body.appendChild(popupDiv);
        
        return storyDiv;
    }
    
    createNoResultsElement() {
        const noResultsDiv = document.createElement('div');
        noResultsDiv.className = 'no-results';
        noResultsDiv.innerHTML = `
            <div class="no-results-content">
                <h3>No stories found</h3>
                <p>We couldn't find any stories matching your current filters.</p>
                <p>Try adjusting your search criteria or <button type="button" onclick="window.archiveFilters.clearAllFilters()" class="clear-link">clear all filters</button> to explore the full archive.</p>
            </div>
        `;
        
        return noResultsDiv;
    }
    
    updateFilterCount() {
        const total = this.filterData ? this.filterData.stories.length : 0;
        const filtered = this.filteredStories.length;
        const hasActiveFilters = Object.values(this.currentFilters).some(f => f.length > 0);

        // Update the heading text based on whether filters are active
        if (this.elements.archiveHeading) {
            const headingPrefix = hasActiveFilters
                ? 'Filtered items from the archive ('
                : 'All items from the archive (';
            // Detach the span first, then clear text, then re-attach
            const span = this.elements.archiveHeading.querySelector('[data-el="total-count"]');
            if (span) span.remove();
            this.elements.archiveHeading.textContent = headingPrefix;
            if (span) this.elements.archiveHeading.appendChild(span);
            this.elements.archiveHeading.appendChild(document.createTextNode(')'));
        }

        // Update the main total count in the header
        if (this.elements.totalCount) {
            this.elements.totalCount.textContent = filtered;
        }
        
        // Update the filter count display
        if (this.elements.filterCount) {
            if (filtered === total) {
                this.elements.filterCount.textContent = '';
            } else {
                this.elements.filterCount.textContent = `${filtered} of ${total} stories`;
            }
        }
    }
    
    updateActiveFiltersDisplay() {
        if (!this.elements.activeFilters || !this.elements.activeFiltersList) return;
        
        const hasActiveFilters = Object.values(this.currentFilters).some(filters => filters.length > 0);
        
        if (!hasActiveFilters) {
            this.elements.activeFilters.style.display = 'none';
            return;
        }
        
        this.elements.activeFilters.style.display = 'flex';
        this.elements.activeFiltersList.innerHTML = '';
        
        // Add theme filters
        this.currentFilters.themes.forEach(theme => {
            this.elements.activeFiltersList.appendChild(
                this.createActiveFilterTag(theme, 'themes')
            );
        });
        
        // Add type filters
        this.currentFilters.types.forEach(type => {
            this.elements.activeFiltersList.appendChild(
                this.createActiveFilterTag(type, 'types')
            );
        });
        
        // Add weather filters
        this.currentFilters.weather.forEach(weather => {
            this.elements.activeFiltersList.appendChild(
                this.createActiveFilterTag(weather, 'weather')
            );
        });

        // Add what was/is/if filters
        this.currentFilters.whatWasIsIf.forEach(wwii => {
            this.elements.activeFiltersList.appendChild(
                this.createActiveFilterTag(wwii, 'whatWasIsIf')
            );
        });

        // Add scale permanence filters
        this.currentFilters.scalePermanence.forEach(sp => {
            this.elements.activeFiltersList.appendChild(
                this.createActiveFilterTag(sp, 'scalePermanence')
            );
        });

        // Add time period filters
        this.currentFilters.timePeriod.forEach(tp => {
            this.elements.activeFiltersList.appendChild(
                this.createActiveFilterTag(tp, 'timePeriod')
            );
        });
    }
    
    createActiveFilterTag(value, filterType) {
        const tag = document.createElement('button');
        tag.className = 'active-filters__tag tag';
        
        // Find the original color for this filter option
        let color = '#666666'; // Default color
        if (this.filterData) {
            const filterOptions = this.filterData[filterType];
            const option = filterOptions.find(opt => opt.title === value);
            if (option) {
                color = option.color || this.getDefaultColor(option.title, filterType);
            }
        }
        
        // Apply the same color as the dropdown option
        tag.style.backgroundColor = color;
        tag.style.color = '#ffffff';
        
        tag.innerHTML = `
            ${value}
            <span class="tag__remove" aria-label="Remove filter">×</span>
        `;
        
        tag.addEventListener('click', () => this.removeFilter(value, filterType));
        
        return tag;
    }
    
    removeFilter(value, filterType) {
        const index = this.currentFilters[filterType].indexOf(value);
        if (index > -1) {
            this.currentFilters[filterType].splice(index, 1);
        }
        
        this.updateDropdownDisplay();
        this.applyFilters();
        this.updateURL();
        this.updateActiveFiltersDisplay();
    }
    
    clearAllFilters() {
        this.currentFilters = {
            themes: [],
            types: [],
            weather: [],
            whatWasIsIf: [],
            scalePermanence: [],
            timePeriod: []
        };
        
        this.updateDropdownDisplay();
        this.closeAllDropdowns();
        this.applyFilters();
        this.updateURL();
        this.updateActiveFiltersDisplay();
    }
    
    /**
     * Load more stories when scrolling near the bottom
     * 
     * This is "lazy loading" or "infinite scroll". Instead of showing all stories
     * at once (which would be slow with hundreds of stories), we show 20 at a time.
     * 
     * When you scroll near the bottom, this function automatically loads the next
     * batch. The user never sees a "Load more" button - it just happens smoothly
     * as they scroll.
     * 
     * The 200px buffer means we start loading before you actually reach the bottom,
     * so stories appear just as you need them.
     */
    handleScroll() {
        if (!this.elements.storiesContainer) return;
        
        const container = this.elements.storiesContainer;
        const scrollTop = window.pageYOffset || document.documentElement.scrollTop;
        const windowHeight = window.innerHeight;
        const containerBottom = container.offsetTop + container.offsetHeight;
        
        // Are we within 200 pixels of the bottom?
        if (scrollTop + windowHeight >= containerBottom - 200) {
            const currentlyDisplayed = container.children.length;
            const totalFiltered = this.filteredStories.length;
            
            // If there are more stories to show, load the next batch
            if (currentlyDisplayed < totalFiltered) {
                this.renderStoriesBatch(currentlyDisplayed, this.storiesPerPage);
            }
        }
    }
    
    updateURL() {
        const params = new URLSearchParams();
        
        if (this.currentFilters.themes.length > 0) {
            params.set('themes', this.currentFilters.themes.join(','));
        }
        
        if (this.currentFilters.types.length > 0) {
            params.set('types', this.currentFilters.types.join(','));
        }
        
        if (this.currentFilters.weather.length > 0) {
            params.set('weather', this.currentFilters.weather.join(','));
        }

        if (this.currentFilters.whatWasIsIf.length > 0) {
            params.set('whatWasIsIf', this.currentFilters.whatWasIsIf.join(','));
        }

        if (this.currentFilters.scalePermanence.length > 0) {
            params.set('scalePermanence', this.currentFilters.scalePermanence.join(','));
        }

        if (this.currentFilters.timePeriod.length > 0) {
            params.set('timePeriod', this.currentFilters.timePeriod.join(','));
        }

        if (this.currentSort && this.currentSort !== 'dateExperienced') {
            params.set('sort', this.currentSort);
        }

        const newURL = params.toString() ? `?${params.toString()}` : window.location.pathname;
        window.history.pushState(null, '', newURL);
    }
    
    loadFiltersFromURL() {
        const params = new URLSearchParams(window.location.search);
        
        this.currentFilters.themes = params.get('themes') ? params.get('themes').split(',') : [];
        this.currentFilters.types = params.get('types') ? params.get('types').split(',') : [];
        this.currentFilters.weather = params.get('weather') ? params.get('weather').split(',') : [];
        this.currentFilters.whatWasIsIf = params.get('whatWasIsIf') ? params.get('whatWasIsIf').split(',') : [];
        this.currentFilters.scalePermanence = params.get('scalePermanence') ? params.get('scalePermanence').split(',') : [];
        this.currentFilters.timePeriod = params.get('timePeriod') ? params.get('timePeriod').split(',') : [];

        // Restore sort from URL
        const sortParam = params.get('sort');
        if (sortParam && ['dateExperienced', 'dateCreated', 'random'].includes(sortParam)) {
            this.currentSort = sortParam;
            const labels = { dateExperienced: 'Date experienced', dateCreated: 'Date created', random: 'Random' };
            if (this.elements.sortText) {
                this.elements.sortText.textContent = labels[sortParam];
            }
            if (this.elements.sortContent) {
                this.elements.sortContent.querySelectorAll('.sort-option').forEach(opt => {
                    opt.classList.toggle('selected', opt.dataset.sort === sortParam);
                });
            }
        }

        // If any "more" filters are active from URL, auto-expand the more filters section
        const hasMoreFilterParams = this.currentFilters.scalePermanence.length > 0 ||
            this.currentFilters.timePeriod.length > 0;
        if (hasMoreFilterParams && this.elements.moreFiltersContainer) {
            this.elements.moreFiltersContainer.style.display = '';
            if (this.elements.moreFiltersToggle) {
                this.elements.moreFiltersToggle.classList.add('open');
            }
        }

        // Update dropdown displays
        if (this.filterData) {
            this.updateDropdownDisplay();
            this.applyFilters();
            this.updateActiveFiltersDisplay();
        }
    }
    
    showError(message) {
        const errorDiv = document.createElement('div');
        errorDiv.className = 'filter-error';
        errorDiv.style.cssText = `
            background: #ffebee;
            border: 2px solid #f44336;
            padding: 1rem;
            margin: 1rem 0;
            font-family: Arial, sans-serif;
            color: #c62828;
        `;
        errorDiv.textContent = message;
        
        if (this.elements.storiesContainer && this.elements.storiesContainer.parentNode) {
            this.elements.storiesContainer.parentNode.insertBefore(errorDiv, this.elements.storiesContainer);
        }
    }
    
    // Utility function to debounce scroll events
    debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }

    /**
     * Sort an array of strings alphabetically
     *
     * Returns a new sorted array without modifying the original.
     * Used to ensure tags are displayed in a consistent order.
     */
    sortAlphabetically(items) {
        if (!items || items.length === 0) {
            return [];
        }

        // Create a copy and sort it
        const sorted = items.slice();
        sorted.sort();
        return sorted;
    }
}

// Initialize the archive filters when the DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.archiveFilters = new ArchiveFilters();
});
