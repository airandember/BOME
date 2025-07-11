<script lang="ts">
	import { onMount } from 'svelte';
	import Navigation from '$lib/components/Navigation.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import AdDisplay from '$lib/components/AdDisplay.svelte';

	interface Event {
		id: number;
		title: string;
		description: string;
		date: string;
		time: string;
		location: string;
		type: 'conference' | 'seminar' | 'exhibition' | 'workshop';
		price: number;
		capacity: number;
		registrations: number;
		image: string;
		organizer: string;
		tags: string[];
		featured: boolean;
	}

	let events: Event[] = [];
	let filteredEvents: Event[] = [];
	let loading = true;
	let searchQuery = '';
	let selectedType = '';
	let selectedLocation = '';
	let dateFilter = '';
	let currentPage = 1;
	let eventsPerPage = 9;

	// Mock data for events
	const mockEvents: Event[] = [
		{
			id: 1,
			title: "Book of Mormon Archaeological Conference 2024",
			description: "Join leading archaeologists and researchers as they present the latest findings supporting Book of Mormon geography and civilizations.",
			date: "2024-04-15",
			time: "09:00 AM",
			location: "Salt Lake City, UT",
			type: "conference",
			price: 150,
			capacity: 500,
			registrations: 342,
			image: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
			organizer: "BOME Research Institute",
			tags: ["archaeology", "research", "Book of Mormon"],
			featured: true
		},
		{
			id: 2,
			title: "DNA and Ancient Migrations Seminar",
			description: "Explore the latest DNA research and its implications for understanding ancient American populations.",
			date: "2024-05-20",
			time: "02:00 PM",
			location: "Provo, UT",
			type: "seminar",
			price: 75,
			capacity: 200,
			registrations: 156,
			image: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
			organizer: "Genetic Research Foundation",
			tags: ["DNA", "genetics", "ancient migrations"],
			featured: false
		},
		{
			id: 3,
			title: "Ancient Civilizations Exhibition",
			description: "Interactive exhibition featuring artifacts, models, and multimedia presentations of ancient American civilizations.",
			date: "2024-06-01",
			time: "10:00 AM",
			location: "Mesa, AZ",
			type: "exhibition",
			price: 25,
			capacity: 1000,
			registrations: 687,
			image: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
			organizer: "Desert Museum",
			tags: ["exhibition", "artifacts", "civilizations"],
			featured: true
		},
		{
			id: 4,
			title: "Mesoamerican Studies Workshop",
			description: "Hands-on workshop for students and researchers interested in Mesoamerican archaeology and culture.",
			date: "2024-07-10",
			time: "08:00 AM",
			location: "Guatemala City, GT",
			type: "workshop",
			price: 200,
			capacity: 50,
			registrations: 23,
			image: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
			organizer: "International Archaeological Society",
			tags: ["Mesoamerica", "workshop", "archaeology"],
			featured: false
		},
		{
			id: 5,
			title: "Book of Mormon Geography Symposium",
			description: "Comprehensive symposium exploring various theories and evidence for Book of Mormon geography.",
			date: "2024-08-25",
			time: "09:30 AM",
			location: "Independence, MO",
			type: "conference",
			price: 125,
			capacity: 300,
			registrations: 189,
			image: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
			organizer: "Geography Research Center",
			tags: ["geography", "theories", "symposium"],
			featured: true
		},
		{
			id: 6,
			title: "Ancient Metallurgy Seminar",
			description: "Examining ancient metalworking techniques and their presence in pre-Columbian America.",
			date: "2024-09-15",
			time: "01:00 PM",
			location: "Denver, CO",
			type: "seminar",
			price: 85,
			capacity: 150,
			registrations: 98,
			image: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
			organizer: "Metallurgy Institute",
			tags: ["metallurgy", "ancient technology", "pre-Columbian"],
			featured: false
		}
	];

	$: totalPages = Math.ceil(filteredEvents.length / eventsPerPage);
	$: paginatedEvents = filteredEvents.slice(
		(currentPage - 1) * eventsPerPage,
		currentPage * eventsPerPage
	);
	$: featuredEvents = events.filter(event => event.featured);

	onMount(() => {
		// Simulate loading
		setTimeout(() => {
			events = mockEvents;
			applyFilters();
			loading = false;
		}, 1000);
	});

	function applyFilters() {
		let result = [...events];

		// Search filter
		if (searchQuery.trim()) {
			result = result.filter(event =>
				event.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
				event.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
				event.tags.some(tag => tag.toLowerCase().includes(searchQuery.toLowerCase()))
			);
		}

		// Type filter
		if (selectedType) {
			result = result.filter(event => event.type === selectedType);
		}

		// Location filter
		if (selectedLocation) {
			result = result.filter(event =>
				event.location.toLowerCase().includes(selectedLocation.toLowerCase())
			);
		}

		// Date filter
		if (dateFilter) {
			const filterDate = new Date(dateFilter);
			result = result.filter(event => {
				const eventDate = new Date(event.date);
				return eventDate >= filterDate;
			});
		}

		// Sort by date
		result.sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime());

		filteredEvents = result;
		currentPage = 1;
	}

	function handleSearch() {
		applyFilters();
	}

	function clearFilters() {
		searchQuery = '';
		selectedType = '';
		selectedLocation = '';
		dateFilter = '';
		applyFilters();
	}

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}

	function getEventTypeLabel(type: string): string {
		const labels: Record<string, string> = {
			conference: 'Conference',
			seminar: 'Seminar',
			exhibition: 'Exhibition',
			workshop: 'Workshop'
		};
		return labels[type] || type;
	}

	function getAvailabilityStatus(event: any): { status: string; color: string; text: string; canRegister: boolean } {
		const spotsLeft = event.capacity - event.registrations;
		
		if (spotsLeft <= 0) {
			return { 
				status: 'full',
				color: 'red', 
				text: 'Full',
				canRegister: false
			};
		} else if (spotsLeft <= 5) {
			return { 
				status: 'limited',
				color: 'orange', 
				text: 'Limited spots',
				canRegister: true
			};
		} else {
			return { 
				status: 'available',
				color: 'green', 
				text: 'Available',
				canRegister: true
			};
		}
	}

	function handleLearnMore(event: any) {
		// Navigate to event details or show modal
		console.log('Learn more about:', event.title);
	}
	
	function handleRegister(event: any) {
		// Handle registration
		console.log('Register for:', event.title);
	}
	
	function formatEventDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString('en-US', {
			weekday: 'short',
			month: 'short',
			day: 'numeric'
		});
	}
</script>

<svelte:head>
	<title>Events & Exhibitions - BOME | Book of Mormon Evidence</title>
	<meta name="description" content="Join conferences, seminars, workshops, and exhibitions focused on Book of Mormon evidence and research." />
	<meta name="keywords" content="Book of Mormon, events, conferences, seminars, exhibitions, workshops, archaeology" />
</svelte:head>

<Navigation />

<main class="events-container">
	<!-- Hero Section -->
	<section class="hero-section">
		<div class="hero-content">
			<h1 class="hero-title">Events & Exhibitions</h1>
			<p class="hero-description">
				Join conferences, seminars, workshops, and exhibitions focused on Book of Mormon evidence, 
				archaeology, and research. Connect with experts and fellow enthusiasts.
			</p>
		</div>
		
		<!-- Ad Placement: Header Banner -->
		<div class="ad-placement">
			<AdDisplay placement="events-header" />
		</div>
	</section>

	<!-- Featured Events -->
	{#if featuredEvents.length > 0}
		<section class="featured-section">
			<h2 class="section-title">Featured Events</h2>
			<div class="featured-grid">
				{#each featuredEvents as event}
					<article class="featured-card">
						<div class="featured-image">
							<img src={event.image} alt={event.title} />
							<div class="featured-overlay">
								<span class="featured-badge">Featured</span>
								<span class="type-badge {event.type}">{getEventTypeLabel(event.type)}</span>
							</div>
						</div>
						<div class="featured-content">
							<div class="event-meta">
								<span class="event-date">{formatDate(event.date)}</span>
								<span class="event-time">{event.time}</span>
							</div>
							<h3 class="featured-title">{event.title}</h3>
							<p class="featured-excerpt">{event.description}</p>
							<div class="event-details">
								<div class="location">📍 {event.location}</div>
								<div class="organizer">🏢 {event.organizer}</div>
								<div class="price">${event.price}</div>
							</div>
							<div class="availability">
								<span class="availability-status {getAvailabilityStatus(event).color}">
									{getAvailabilityStatus(event).status}
								</span>
								<span class="spots-left">
									{event.capacity - event.registrations} spots left
								</span>
							</div>
						</div>
					</article>
				{/each}
			</div>
		</section>
	{/if}

	<!-- Ad Placement: Between Featured and Filters -->
	<div class="ad-placement">
		<AdDisplay placement="events-mid" />
	</div>

	<!-- Filters Section -->
	<section class="filters-section">
		<div class="filters-container">
			<div class="search-bar">
				<input
					type="text"
					placeholder="Search events..."
					bind:value={searchQuery}
					on:keydown={(e) => e.key === 'Enter' && handleSearch()}
					aria-label="Search events"
				/>
				<button class="btn-primary" on:click={handleSearch}>
					🔍 Search
				</button>
			</div>

			<div class="filter-controls">
				<select bind:value={selectedType} on:change={applyFilters} aria-label="Filter by type">
					<option value="">All Types</option>
					<option value="conference">Conferences</option>
					<option value="seminar">Seminars</option>
					<option value="exhibition">Exhibitions</option>
					<option value="workshop">Workshops</option>
				</select>

				<input
					type="text"
					placeholder="Location..."
					bind:value={selectedLocation}
					on:input={applyFilters}
					aria-label="Filter by location"
				/>

				<input
					type="date"
					bind:value={dateFilter}
					on:change={applyFilters}
					aria-label="Filter by date"
				/>

				<button class="btn-secondary" on:click={clearFilters}>
					Clear Filters
				</button>
			</div>
		</div>
	</section>

	<!-- Events Grid -->
	<section class="events-section" id="events-section">
		{#if loading}
			<div class="loading-state">
				<div class="loading-spinner"></div>
				<p>Loading events...</p>
			</div>
		{:else if filteredEvents.length === 0}
			<div class="empty-state">
				<div class="empty-icon">📅</div>
				<h3>No events found</h3>
				<p>Try adjusting your search or filters</p>
				<button class="btn-primary" on:click={clearFilters}>
					Clear All Filters
				</button>
			</div>
		{:else}
			<div class="events-grid">
				{#each paginatedEvents as event}
					<div class="event-card glass">
						<div class="event-image">
							<img src={event.image} alt={event.title} />
							<div class="event-category">
								<span class="category-badge category-{event.type}">
									{getEventTypeLabel(event.type)}
								</span>
							</div>
						</div>
						
						<div class="event-content">
							<div class="event-header">
								<h3>{event.title}</h3>
								<div class="event-meta">
									<div class="event-date">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
											<line x1="16" y1="2" x2="16" y2="6"></line>
											<line x1="8" y1="2" x2="8" y2="6"></line>
											<line x1="3" y1="10" x2="21" y2="10"></line>
										</svg>
										<span>{formatEventDate(event.date)}</span>
									</div>
									
									<div class="event-location">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"></path>
											<circle cx="12" cy="10" r="3"></circle>
										</svg>
										<span>{event.location}</span>
									</div>
								</div>
							</div>
							
							<p class="event-description">{event.description}</p>
							
							<div class="event-details">
								<div class="event-price">
									<span class="price-label">Price:</span>
									<span class="price-value">
										{event.price > 0 ? `$${event.price}` : 'Free'}
									</span>
								</div>
								
								<div class="availability">
									<span class="availability-status {getAvailabilityStatus(event).color}">
										{getAvailabilityStatus(event).status}
									</span>
									<span class="spots-left">
										{event.capacity - event.registrations} spots left
									</span>
								</div>
							</div>
							
							<div class="event-actions">
								<button class="btn btn-outline" on:click={() => handleLearnMore(event)}>
									Learn More
								</button>
								<button 
									class="btn btn-primary" 
									on:click={() => handleRegister(event)}
									disabled={!getAvailabilityStatus(event).canRegister}
								>
									{getAvailabilityStatus(event).canRegister ? 'Register' : 'Full'}
								</button>
							</div>
						</div>
					</div>
				{/each}
			</div>

			<!-- Pagination -->
			{#if totalPages > 1}
				<div class="pagination">
					<button 
						class="page-btn" 
						disabled={currentPage === 1}
						on:click={() => currentPage = Math.max(1, currentPage - 1)}
					>
						Previous
					</button>
					
					{#each Array(totalPages) as _, i}
						<button 
							class="page-btn {currentPage === i + 1 ? 'active' : ''}"
							on:click={() => currentPage = i + 1}
						>
							{i + 1}
						</button>
					{/each}
					
					<button 
						class="page-btn" 
						disabled={currentPage === totalPages}
						on:click={() => currentPage = Math.min(totalPages, currentPage + 1)}
					>
						Next
					</button>
				</div>
			{/if}
		{/if}
	</section>

	<!-- Ad Placement: Footer -->
	<div class="ad-placement">
		<AdDisplay placement="events-footer" />
	</div>
</main>

<Footer />

<style>
	.events-container {
		min-height: 100vh;
		background: var(--bg-primary);
		color: var(--text-primary);
		padding-top: 70px;
		width: 95vw;
		max-width: none;
		margin: 0 auto;
		padding-left: 1rem;
		padding-right: 1rem;
	}

	.hero-section {
		background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
		color: white;
		padding: 4rem 2rem;
		text-align: center;
		position: relative;
		overflow: hidden;
	}

	.hero-section::before {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><circle cx="20" cy="20" r="2" fill="white" opacity="0.1"/><circle cx="80" cy="40" r="1.5" fill="white" opacity="0.1"/><circle cx="40" cy="70" r="1" fill="white" opacity="0.1"/><circle cx="70" cy="20" r="1.5" fill="white" opacity="0.1"/></svg>');
		animation: float 20s linear infinite;
	}

	@keyframes float {
		0% { transform: translateY(0); }
		100% { transform: translateY(-100px); }
	}

	.hero-content {
		text-align: center;
		margin-bottom: 2rem;
		width: 100%;
		max-width: none;
	}

	.hero-title {
		font-size: 3.5rem;
		font-weight: 800;
		margin-bottom: 1.5rem;
		text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.3);
	}

	.hero-description {
		font-size: 1.2rem;
		opacity: 0.9;
		line-height: 1.6;
		margin-bottom: 2rem;
	}

	/* Featured Section */
	.featured-section {
		padding: 4rem 2rem;
		background: var(--bg-secondary);
		width: 100%;
		max-width: none;
	}

	.section-title {
		font-size: 2.5rem;
		font-weight: 700;
		text-align: center;
		margin-bottom: 3rem;
		color: var(--text-primary);
	}

	.featured-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
		gap: 2rem;
		width: 100%;
		max-width: none;
	}

	.featured-card {
		background: var(--card-bg);
		border-radius: 20px;
		overflow: hidden;
		box-shadow: 
			8px 8px 16px var(--shadow-dark),
			-8px -8px 16px var(--shadow-light);
		transition: transform 0.3s ease, box-shadow 0.3s ease;
	}

	.featured-card:hover {
		transform: translateY(-5px);
		box-shadow: 
			12px 12px 24px var(--shadow-dark),
			-12px -12px 24px var(--shadow-light);
	}

	.featured-image {
		position: relative;
		height: 200px;
		overflow: hidden;
	}

	.featured-image img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		transition: transform 0.3s ease;
	}

	.featured-card:hover .featured-image img {
		transform: scale(1.05);
	}

	.featured-overlay {
		position: absolute;
		top: 1rem;
		left: 1rem;
		right: 1rem;
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
	}

	.featured-badge {
		background: linear-gradient(135deg, #ff6b6b, #ee5a24);
		color: white;
		padding: 0.5rem 1rem;
		border-radius: 20px;
		font-size: 0.8rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.type-badge {
		padding: 0.5rem 1rem;
		border-radius: 20px;
		font-size: 0.8rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		color: white;
	}

	.type-badge.conference { background: linear-gradient(135deg, #4834d4, #686de0); }
	.type-badge.seminar { background: linear-gradient(135deg, #00d2d3, #54a0ff); }
	.type-badge.exhibition { background: linear-gradient(135deg, #ff9ff3, #f368e0); }
	.type-badge.workshop { background: linear-gradient(135deg, #ff9f43, #feca57); }

	.featured-content {
		padding: 2rem;
	}

	.event-meta {
		display: flex;
		gap: 1rem;
		margin-bottom: 1rem;
		font-size: 0.9rem;
		color: var(--text-secondary);
	}

	.featured-title {
		font-size: 1.3rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 1rem;
		line-height: 1.4;
	}

	.featured-excerpt {
		color: var(--text-secondary);
		line-height: 1.6;
		margin-bottom: 1.5rem;
	}

	.event-details {
		display: grid;
		gap: 0.5rem;
		margin-bottom: 1rem;
		font-size: 0.9rem;
	}

	.location, .organizer {
		color: var(--text-secondary);
	}

	.price {
		font-size: 1.2rem;
		font-weight: 700;
		color: var(--primary-color);
	}

	.availability {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.75rem;
		background: var(--bg-secondary);
		border-radius: 10px;
		margin-top: 1rem;
	}

	.availability-status {
		font-weight: 600;
		font-size: 0.9rem;
	}

	.availability-status.success { color: #27ae60; }
	.availability-status.warning { color: #f39c12; }
	.availability-status.danger { color: #e74c3c; }

	.spots-left {
		font-size: 0.8rem;
		color: var(--text-secondary);
	}

	/* Filters Section */
	.filters-section {
		background: var(--bg-secondary);
		padding: 2rem;
		margin: 2rem 0;
	}

	.filters-container {
		width: 100%;
		max-width: none;
		margin: 0 auto 2rem auto;
		background: var(--surface-color);
		border-radius: 12px;
		padding: 1.5rem;
		box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
	}

	.search-bar {
		display: flex;
		gap: 1rem;
		margin-bottom: 1.5rem;
	}

	.search-bar input {
		flex: 1;
		padding: 0.75rem 1rem;
		border: none;
		border-radius: 12px;
		background: var(--input-bg);
		color: var(--text-primary);
		font-size: 1rem;
		box-shadow: 
			inset 2px 2px 4px var(--shadow-dark),
			inset -2px -2px 4px var(--shadow-light);
	}

	.filter-controls {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
		align-items: center;
	}

	.filter-controls select,
	.filter-controls input {
		padding: 0.75rem 1rem;
		border: none;
		border-radius: 12px;
		background: var(--input-bg);
		color: var(--text-primary);
		font-size: 1rem;
		box-shadow: 
			inset 2px 2px 4px var(--shadow-dark),
			inset -2px -2px 4px var(--shadow-light);
	}

	/* Events Grid */
	.events-section {
		padding: 4rem 2rem;
		width: 100%;
		max-width: none;
	}

	.events-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
		gap: 2rem;
		margin-bottom: 3rem;
		width: 100%;
		max-width: none;
	}

	.event-card {
		background: var(--card-bg);
		border-radius: 20px;
		overflow: hidden;
		box-shadow: 
			8px 8px 16px var(--shadow-dark),
			-8px -8px 16px var(--shadow-light);
		transition: transform 0.3s ease, box-shadow 0.3s ease;
	}

	.event-card:hover {
		transform: translateY(-3px);
		box-shadow: 
			12px 12px 24px var(--shadow-dark),
			-12px -12px 24px var(--shadow-light);
	}

	.event-image {
		position: relative;
		height: 180px;
		overflow: hidden;
	}

	.event-image img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.event-overlay {
		position: absolute;
		top: 1rem;
		right: 1rem;
	}

	.event-content {
		padding: 1.5rem;
	}

	.event-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 1rem;
	}

	.event-title {
		font-size: 1.2rem;
		font-weight: 600;
		color: var(--text-primary);
		line-height: 1.4;
		flex: 1;
		margin-right: 1rem;
	}

	.event-price {
		font-size: 1.1rem;
		font-weight: 700;
		color: var(--primary-color);
		white-space: nowrap;
	}

	.event-description {
		color: var(--text-secondary);
		line-height: 1.5;
		margin-bottom: 1rem;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	.event-info {
		display: grid;
		gap: 0.5rem;
		margin-bottom: 1rem;
	}

	.info-item {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.9rem;
		color: var(--text-secondary);
	}

	.info-icon {
		font-size: 1rem;
	}

	.event-tags {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		margin-bottom: 1.5rem;
	}

	.tag {
		background: var(--bg-secondary);
		color: var(--text-secondary);
		padding: 0.25rem 0.75rem;
		border-radius: 20px;
		font-size: 0.8rem;
		font-weight: 500;
	}

	.event-footer {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding-top: 1rem;
		border-top: 1px solid var(--border-color);
	}

	.availability-info {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.register-btn {
		padding: 0.75rem 1.5rem;
		font-weight: 600;
	}

	/* Buttons */
	.btn-primary {
		background: linear-gradient(135deg, var(--primary-color), var(--accent-color));
		color: white;
		border: none;
		padding: 0.75rem 1.5rem;
		border-radius: 12px;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.3s ease;
		box-shadow: 
			4px 4px 8px var(--shadow-dark),
			-4px -4px 8px var(--shadow-light);
	}

	.btn-primary:hover {
		transform: translateY(-2px);
		box-shadow: 
			6px 6px 12px var(--shadow-dark),
			-6px -6px 12px var(--shadow-light);
	}

	.btn-secondary {
		background: var(--bg-secondary);
		color: var(--text-primary);
		border: none;
		padding: 0.75rem 1.5rem;
		border-radius: 12px;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.3s ease;
		box-shadow: 
			4px 4px 8px var(--shadow-dark),
			-4px -4px 8px var(--shadow-light);
	}

	.btn-secondary:hover {
		transform: translateY(-2px);
		box-shadow: 
			6px 6px 12px var(--shadow-dark),
			-6px -6px 12px var(--shadow-light);
	}

	/* Pagination */
	.pagination {
		display: flex;
		justify-content: center;
		gap: 0.5rem;
		margin-top: 2rem;
	}

	.page-btn {
		padding: 0.75rem 1rem;
		border: none;
		background: var(--bg-secondary);
		color: var(--text-primary);
		border-radius: 8px;
		cursor: pointer;
		transition: all 0.3s ease;
		font-weight: 500;
	}

	.page-btn:hover:not(:disabled) {
		background: var(--primary-color);
		color: white;
	}

	.page-btn.active {
		background: var(--primary-color);
		color: white;
	}

	.page-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	/* Loading and Empty States */
	.loading-state,
	.empty-state {
		text-align: center;
		padding: 4rem 2rem;
	}

	.loading-spinner {
		width: 40px;
		height: 40px;
		border: 4px solid var(--bg-secondary);
		border-top: 4px solid var(--primary-color);
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin: 0 auto 1rem;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.empty-icon {
		font-size: 4rem;
		margin-bottom: 1rem;
	}

	.empty-state h3 {
		font-size: 1.5rem;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.empty-state p {
		color: var(--text-secondary);
		margin-bottom: 2rem;
	}

	/* Ad Placements */
	.ad-placement {
		margin: 2rem 0;
		display: flex;
		justify-content: center;
	}

	/* Responsive Design */
	@media (max-width: 1200px) {
		.events-grid {
			grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
			gap: 1.5rem;
		}

		.featured-events {
			grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
			gap: 1.5rem;
		}

		.featured-grid {
			grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
			gap: 1.5rem;
		}
	}

	@media (max-width: 992px) {
		.events-container {
			width: 98vw;
			padding-left: 0.5rem;
			padding-right: 0.5rem;
		}

		.events-grid {
			grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
			gap: 1rem;
		}

		.featured-events {
			grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
			gap: 1rem;
		}

		.featured-grid {
			grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
			gap: 1rem;
		}
	}

	@media (max-width: 768px) {
		.events-grid {
			grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
		}

		.featured-events {
			grid-template-columns: 1fr;
		}

		.featured-grid {
			grid-template-columns: 1fr;
		}

		.filters-container {
			padding: 1rem;
		}
	}

	@media (max-width: 480px) {
		.events-container {
			width: 100vw;
			padding-left: 0.25rem;
			padding-right: 0.25rem;
		}

		.events-grid {
			grid-template-columns: 1fr;
		}

		.filters-container {
			padding: 0.75rem;
		}
	}
</style> 