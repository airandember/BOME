<script lang="ts">
	import { createEventDispatcher } from 'svelte';

	export let activeTab: 'subscribers' | 'non-subscribers' = 'subscribers';
	export let subscriberCount = 0;
	export let nonSubscriberCount = 0;

	const dispatch = createEventDispatcher<{
		tabChange: { tab: 'subscribers' | 'non-subscribers' };
	}>();

	function handleTabChange(tab: 'subscribers' | 'non-subscribers') {
		dispatch('tabChange', { tab });
	}
</script>

<div class="tabs">
	<button 
		class="tab {activeTab === 'subscribers' ? 'active' : ''}" 
		on:click={() => handleTabChange('subscribers')}
	>
		Subscribers ({subscriberCount})
	</button>
	<button 
		class="tab {activeTab === 'non-subscribers' ? 'active' : ''}" 
		on:click={() => handleTabChange('non-subscribers')}
	>
		Non-Subscribers ({nonSubscriberCount})
	</button>
</div>

<style>
	.tabs {
		display: flex;
		background: var(--bg-secondary);
		border-radius: 12px;
		padding: 0.25rem;
		margin-bottom: 2rem;
		justify-content: space-around;
	}

	.tab {
		flex: 1;
		padding: 0.75rem 1.5rem;
		border: none;
		background: none;
		color: var(--text-secondary);
		font-weight: 500;
		cursor: pointer;
		border-radius: 8px;
		transition: all 0.3s ease;
		max-width: 400px;
		height: 65px;
		border-radius: 50px;
		
		background: linear-gradient(145deg, var(--secondary), var(--tertiary));
		box-shadow: inset 2px 2px 7px var(--bg-quinary),
					inset -2px -2px 7px #ffffff;
		transition: box-shadow 1s ease-in-out, transform 0.2s ease-in-out;
	}

	.tab:hover {
		color: var(--text-primary);
		background: var(--bg-tertiary);
		transform: translateY(-2px);
		box-shadow: 5px 5px 15px var(--bg-quaternary),
					-5px -5px 15px var(--bg-secondary);
	}

	.tab:active {
		transform: translateY(0);
		box-shadow: inset 3px 3px 10px var(--bg-quinary),
					inset -3px -3px 10px #ffffff;
	}

	.tab.active {
		background: var(--bg-dark);
		color: var(--text-primary);
		background: linear-gradient(145deg, var(--secondary), var(--tertiary));
		box-shadow:  5px 5px 10px var(--bg-quaternary),
              		-5px -5px 10px var(--bg-secondary);
	}

	@media (max-width: 768px) {
		.tabs {
			flex-direction: column;
		}
	}
</style> 