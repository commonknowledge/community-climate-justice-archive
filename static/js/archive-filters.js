/**
 * Archive Filtering System
 * 
 * Provides client-side filtering for the Dudley Climate Justice Archive
 * using vanilla JavaScript with no dependencies.
 */

class ArchiveFilters {
    constructor() {
        this.filterData = null;
        this.currentFilters = {
            themes: [],
            types: [],
            weather: []
        };
        this.filteredStories = [];
        this.currentPage = 0;
        this.storiesPerPage = 20;
        
        this.initializeElements();
        this.loadFilterData();
        this.setupEventListeners();
        this.loadFiltersFromURL();
    }
    
    initializeElements() {
        this.elements = {
            themeDropdown: document.getElementById('theme-dropdown'),
            typeDropdown: document.getElementById('type-dropdown'),
            weatherDropdown: document.getElementById('weather-dropdown'),
            themeButton: document.getElementById('theme-button'),
            typeButton: document.getElementById('type-button'),
            weatherButton: document.getElementById('weather-button'),
            themeContent: document.getElementById('theme-content'),
            typeContent: document.getElementById('type-content'),
            weatherContent: document.getElementById('weather-content'),
            clearFilters: document.getElementById('clear-filters'),
            filterCount: document.getElementById('filter-count'),
            activeFilters: document.getElementById('active-filters'),
            activeFiltersList: document.getElementById('active-filters-list'),
            storiesContainer: document.getElementById('stories-container')
        };
        
        // Verify all elements exist
        for (const [name, element] of Object.entries(this.elements)) {
            if (!element) {
                console.warn(`Filter element not found: ${name}`);
            }
        }
    }
    
    async loadFilterData() {
        try {
            const response = await fetch('/filter-data.json');
            if (!response.ok) {
                throw new Error(`Failed to load filter data: ${response.status}`);
            }
            
            this.filterData = await response.json();
            this.populateFilterDropdowns();
            this.filteredStories = [...this.filterData.stories];
            this.updateFilterCount();
            
        } catch (error) {
            console.error('Error loading filter data:', error);
            this.showError('Failed to load filtering options. Please refresh the page.');
        }
    }
    
    populateFilterDropdowns() {
        if (!this.filterData) return;
        
        // Populate themes
        this.populateDropdown(this.elements.themeContent, this.filterData.themes, 'themes');
        
        // Populate types
        this.populateDropdown(this.elements.typeContent, this.filterData.types, 'types');
        
        // Populate weather
        this.populateDropdown(this.elements.weatherContent, this.filterData.weather, 'weather');
    }
    
    populateDropdown(contentElement, options, filterType) {
        if (!contentElement || !options) return;
        
        // Clear existing content
        contentElement.innerHTML = '';
        
        // Add options sorted by title
        const sortedOptions = [...options].sort((a, b) => a.title.localeCompare(b.title));
        
        sortedOptions.forEach(option => {
            const tagButton = document.createElement('button');
            tagButton.className = 'filter-tag-option';
            
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
        
        // Close dropdowns when clicking outside
        document.addEventListener('click', (e) => {
            if (!e.target.closest('.filter-dropdown')) {
                this.closeAllDropdowns();
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
        
        // Close all dropdowns first
        this.closeAllDropdowns();
        
        // Toggle the clicked dropdown
        if (!isOpen) {
            dropdown.classList.add('open');
        }
    }
    
    closeAllDropdowns() {
        ['theme', 'type', 'weather'].forEach(type => {
            const dropdown = this.elements[`${type}Dropdown`];
            if (dropdown) {
                dropdown.classList.remove('open');
            }
        });
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
        ['themes', 'types', 'weather'].forEach(filterType => {
            const contentElement = this.elements[`${filterType.slice(0, -1)}Content`];
            if (!contentElement) return;
            
            const options = contentElement.querySelectorAll('.filter-tag-option');
            options.forEach(option => {
                const isSelected = this.currentFilters[filterType].includes(option.dataset.value);
                option.classList.toggle('selected', isSelected);
            });
        });
    }
    
    applyFilters() {
        if (!this.filterData) return;
        
        this.filteredStories = this.filterData.stories.filter(story => {
            // Check themes filter
            if (this.currentFilters.themes.length > 0) {
                const hasMatchingTheme = this.currentFilters.themes.some(theme => 
                    story.themes.includes(theme)
                );
                if (!hasMatchingTheme) return false;
            }
            
            // Check types filter
            if (this.currentFilters.types.length > 0) {
                const hasMatchingType = this.currentFilters.types.some(type => 
                    story.types.includes(type)
                );
                if (!hasMatchingType) return false;
            }
            
            // Check weather filter
            if (this.currentFilters.weather.length > 0) {
                const hasMatchingWeather = this.currentFilters.weather.some(weather => 
                    story.weather.includes(weather)
                );
                if (!hasMatchingWeather) return false;
            }
            
            return true;
        });
        
        this.currentPage = 0;
        this.renderStories();
        this.updateFilterCount();
    }
    
    renderStories() {
        if (!this.elements.storiesContainer) return;
        
        // Show loading state
        this.elements.storiesContainer.classList.add('filtering');
        
        // Use a small delay to allow the loading state to show
        setTimeout(() => {
            // Clear existing stories
            this.elements.storiesContainer.innerHTML = '';
            
            if (this.filteredStories.length === 0) {
                // Show friendly no results message
                this.elements.storiesContainer.appendChild(this.createNoResultsElement());
            } else {
                // Render initial batch of stories
                this.renderStoriesBatch(0, this.storiesPerPage);
            }
            
            // Remove loading state
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
        
        if (story.image && story.image.url) {
            storyDiv.innerHTML = `
                <a href="${story.url}">
                    <img 
                        src="${story.image.thumbUrl}" 
                        srcset="${story.image.thumbUrl} 300w, ${story.image.mediumUrl} 800w"
                        sizes="(max-width: 600px) 300px, 800px"
                        alt="${story.image.alt || ''}"
                        loading="lazy">
                </a>
            `;
        } else {
            storyDiv.innerHTML = '<div>No image</div>';
        }
        
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
        if (!this.elements.filterCount) return;
        
        const total = this.filterData ? this.filterData.stories.length : 0;
        const filtered = this.filteredStories.length;
        
        if (filtered === total) {
            this.elements.filterCount.textContent = '';
        } else {
            this.elements.filterCount.textContent = `${filtered} of ${total} stories`;
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
    }
    
    createActiveFilterTag(value, filterType) {
        const tag = document.createElement('button');
        tag.className = 'active-filter-tag';
        
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
        tag.style.color = this.getContrastColor(color);
        
        tag.innerHTML = `
            ${value}
            <span class="active-filter-remove" aria-label="Remove filter">×</span>
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
            weather: []
        };
        
        this.updateDropdownDisplay();
        this.closeAllDropdowns();
        this.applyFilters();
        this.updateURL();
        this.updateActiveFiltersDisplay();
    }
    
    handleScroll() {
        if (!this.elements.storiesContainer) return;
        
        const container = this.elements.storiesContainer;
        const scrollTop = window.pageYOffset || document.documentElement.scrollTop;
        const windowHeight = window.innerHeight;
        const containerBottom = container.offsetTop + container.offsetHeight;
        
        // Check if we're near the bottom and have more stories to load
        if (scrollTop + windowHeight >= containerBottom - 200) {
            const currentlyDisplayed = container.children.length;
            const totalFiltered = this.filteredStories.length;
            
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
        
        const newURL = params.toString() ? `?${params.toString()}` : window.location.pathname;
        window.history.pushState(null, '', newURL);
    }
    
    loadFiltersFromURL() {
        const params = new URLSearchParams(window.location.search);
        
        this.currentFilters.themes = params.get('themes') ? params.get('themes').split(',') : [];
        this.currentFilters.types = params.get('types') ? params.get('types').split(',') : [];
        this.currentFilters.weather = params.get('weather') ? params.get('weather').split(',') : [];
        
        // Update dropdown displays
        if (this.filterData) {
            this.updateDropdownDisplay();
            this.applyFilters();
            this.updateActiveFiltersDisplay();
        }
    }
    
    updateSelectFromFilters(dropdownType, filterValues) {
        // This method is now handled by updateDropdownDisplay()
        // but we keep it for compatibility with loadFiltersFromURL
        if (this.filterData) {
            this.updateDropdownDisplay();
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
}

// Initialize the archive filters when the DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.archiveFilters = new ArchiveFilters();
});
