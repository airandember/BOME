<script lang="ts">
	import { auth } from '$lib/auth';
	import { goto } from '$app/navigation';
	import Navigation from '$lib/components/Navigation.svelte';

	let isAuthenticated = false;
	let openFaq: number | null = null;

	// Subscribe to auth store
	auth.subscribe(state => {
		isAuthenticated = state.isAuthenticated;
	});

	const handleGetStarted = () => {
		if (isAuthenticated) {
			goto('/videos');
		} else {
			goto('/auth/register');
		}
	};

	const handleSubscribe = () => {
		goto('/subscription');
	};

	const toggleFaq = (index: number) => {
		openFaq = openFaq === index ? null : index;
	};
</script>

<svelte:head>
	<title>BOME - Book of Mormon Evidence | Modern Streaming Platform</title>
	<meta name="description" content="Discover compelling evidence for the Book of Mormon through our modern streaming platform. High-quality videos, expert analysis, and exclusive content." />
	<meta name="keywords" content="Book of Mormon, evidence, streaming, LDS, archaeology, history" />
</svelte:head>

<Navigation />

<div class="page-wrapper home-page">
	<!-- Section 1: Hero (Book Close-up) — data-slot for future admin injection -->
	<section class="hero-section book-section" data-slot="hero">
		<div class="zoom-content">
			<div class="book-container">
				<img src="/Images/BOME-Background.webp" alt="Book of Mormon Close-up" class="book-image" />
				<div class="book-glow"></div>
			</div>

			<!-- Hero content slot: headline, copy, CTA — editable from admin -->
			<div class="hero-content-slot navigation-cards" data-slot="hero-content">
				<!--<img src="/Images/bome-evidence-logo.png" alt="Book of Mormon Evidence" class="hero-logo" />-->
				<div class="hero-copy">
					<h1 class="hero-headline">1,600+ Videos. 20 Years of Research.<br>One Streaming Library.</h1>
					<p class="hero-sub">
						The largest collection of Heartland Book of Mormon evidence ever assembled — plus conferences on the Signs of the Times, the Constitution, self-reliance, and health. All in one place. Watch anytime.
					</p>
				</div>
				<a href="/subscription" class="btn btn--gold btn--large subscribe-btn" on:click|preventDefault={handleSubscribe}>
					Subscribe Now
				</a>
				<p class="hero-pricing">
					Just $9.97/mo — or save 20% with an annual plan at $7.97/mo. Cancel anytime.
				</p>
			</div>
		</div>
	</section>

	<!-- Section 2: Problem / Tension — data-slot for future admin injection -->
	<section class="section section--white" data-slot="body-1">
		<div class="container container--narrow text-center tension">
			<h2 class="section-heading" style="color: var(--sales-navy);">You Believe. But the Questions Don't Stop.</h2>
			<p>
				You've felt the Spirit confirm the Book of Mormon is true. That hasn't changed. But the questions keep coming — from your kids, from investigators, from critics online, and sometimes from the quiet part of your own mind that just wants to see the evidence laid out clearly.
			</p>
			<p>
				The problem isn't a lack of evidence. It's that the evidence is scattered across decades of research, buried in conference recordings, and mixed in with material that ranges from careful scholarship to pure speculation. You shouldn't have to sort through all of that alone.
			</p>
			<p>What if someone already did?</p>
		</div>
	</section>

	<!-- Section 3: Library Overview — data-slot for future admin injection -->
	<section class="section section--off-white" data-slot="body-2">
		<div class="container  text-center">
			<div class="container-grove">
				<h2 class="section-heading" style="color: var(--sales-navy);">Five Libraries. One Subscription.</h2>
				<p class="section-subheading">
					Every video from 37 FIRM Foundation Expos and years of independent research — organized into five core categories and growing every quarter.
				</p>
				<div class="categories-grid">
					<div class="category-card">
						<h3>Book of Mormon Evidence</h3>
						<p>The core of the collection. Heartland geography, archaeological findings in ancient North America, DNA research including Haplogroup X, prophetic promises tied to the Land of Promise, and Joseph Smith's own words about where the events of the Book of Mormon took place.</p>
					</div>
					<div class="category-card">
						<h3>Signs of the Times</h3>
						<p>Latter-day prophecy examined through current events and scripture. What the prophets said would happen, what's happening now, and what the scriptures indicate is still ahead. Grounded in text, not in speculation.</p>
					</div>
					<div class="category-card">
						<h3>Constitution</h3>
						<p>America's founding as a covenant land. The prophetic role of the Constitution, the faith of the founders, the scriptural promises attached to this nation, and why it matters today. This isn't political commentary — it's scriptural context.</p>
					</div>
					<div class="category-card">
						<h3>Self-Reliance</h3>
						<p>Practical preparedness and temporal self-sufficiency. Food storage, financial readiness, skill-building, and the principles behind the counsel the prophets have given for decades. Grounded, actionable, no fear-mongering.</p>
					</div>
					<div class="category-card">
						<h3>Health &amp; Wellness</h3>
						<p>Physical stewardship — your body as a temple, taken literally. Practical health topics, nutrition, and wellness principles drawn from both modern research and scriptural guidance.</p>
					</div>
				</div>	
				<p class="library-closing">New content is added regularly from live conferences, events, and ongoing research. The library keeps growing.</p>

			</div>
				<img class="image-grove-left" src="https://assets.cdn.filesafe.space/xar8tXZK2mQ4CKH0W84j/media/68a793c35727bdfd60a74f73.jpeg" alt="Grove1"/>
				<img class="image-grove-right" src="https://assets.cdn.filesafe.space/xar8tXZK2mQ4CKH0W84j/media/68a793c35727bdfd60a74f73.jpeg" alt="Grove2" />
			
		</div>
	</section>

	<!-- Section 4: Stats Banner — data-slot for future admin injection -->
	<section class="section section--navy stats-banner" data-slot="body-3">
		<div class="container">
			<div class="sales-stats-grid">
				<div class="sales-stat-item">
					<div class="stat-item__number">1,600+</div>
					<div class="stat-item__label">Videos in the Library</div>
				</div>
				<div class="sales-stat-item">
					<div class="stat-item__number">20+</div>
					<div class="stat-item__label">Years of Research</div>
				</div>
				<div class="sales-stat-item">
					<div class="stat-item__number">37</div>
					<div class="stat-item__label">FIRM Foundation Expos</div>
				</div>
				<div class="sales-stat-item">
					<div class="stat-item__number">7+</div>
					<div class="stat-item__label">Expert Speakers</div>
				</div>
			</div>
		</div>
	</section>

	<!-- Section 5: Featured Speakers — data-slot for future admin injection -->
	<section class="section section--white" data-slot="body-4">
		<div class="container text-center">
			<h2 class="section-heading" style="color: var(--sales-navy);">The Researchers Behind the Evidence</h2>
			<div class="speakers-grid">
				<div class="speaker-card">
					<div class="speaker-card__icon"><span>RM</span></div>
					<h4>Rod Meldrum</h4>
					<p>Founder of Book of Mormon Evidence. Twenty years of Heartland research, author, and the most-watched presenter in the field.</p>
				</div>
				<div class="speaker-card">
					<div class="speaker-card__icon"><span>HS</span></div>
					<h4>Hannah Stoddard</h4>
					<p>Executive Director of the Joseph Smith Foundation. Researcher, author, and advocate for the prophetic foundations of the Restoration.</p>
				</div>
				<div class="speaker-card">
					<div class="speaker-card__icon"><span>RB</span></div>
					<h4>Russell H. Barlow</h4>
					<p>Founder of TruthSeekers Foundation. Author of <em>Water — Another Testament of Christ</em>. Faith-based science researcher and educator.</p>
				</div>
				<div class="speaker-card">
					<div class="speaker-card__icon"><span>DB</span></div>
					<h4>Don Bradley</h4>
					<p>Historian, researcher, and author. One of the most respected independent scholars working on early Restoration history.</p>
				</div>
				<div class="speaker-card">
					<div class="speaker-card__icon"><span>JB</span></div>
					<h4>Jonah Barnes</h4>
					<p>Researcher and conference presenter. Focused on scriptural geography and Book of Mormon historicity.</p>
				</div>
				<div class="speaker-card">
					<div class="speaker-card__icon"><span>CE</span></div>
					<h4>Carden Ellis</h4>
					<p>Presenter and researcher featured at FIRM Foundation conferences on archaeology and scriptural evidence.</p>
				</div>
				<div class="speaker-card">
					<div class="speaker-card__icon"><span>SJ</span></div>
					<h4>The Stick of Joseph</h4>
					<p>A featured series exploring the prophetic significance of the Book of Mormon as the Stick of Joseph referenced in Ezekiel 37.</p>
				</div>
			</div>
		</div>
	</section>

	<!-- Section 6: Who This Is For — data-slot for future admin injection -->
	<section class="section section--off-white" data-slot="body-5">
		<div class="container container--narrow">
			<h2 class="section-heading text-center" style="color: var(--sales-navy);">This Is For You If...</h2>
			<ul class="for-you-list">
				<li>You study Come, Follow Me every week and want deeper historical and geographic context for what you're reading in the Book of Mormon.</li>
				<li>You've been asked hard questions by your children, by missionaries you support, or by friends outside the faith — and you want clear, evidence-based answers to point them to.</li>
				<li>You believe the Book of Mormon is a real history of real people in a real place — and you want to see what 20 years of Heartland research has uncovered.</li>
				<li>You're tired of critics controlling the conversation. You want to study the evidence yourself, at your own pace, and draw your own conclusions.</li>
				<li>You've watched Rod's YouTube videos and thought, "There has to be more." There is. This is the full library.</li>
			</ul>
			<div class="text-center" style="margin-top: 40px;">
				<a href="/subscription" class="btn btn--charcoal" on:click|preventDefault={handleSubscribe}>Subscribe Now</a>
			</div>
		</div>
	</section>

	<!-- Section 7: Sample Content — data-slot for future admin injection -->
	<section class="section section--navy" data-slot="body-6">
		<div class="container text-center">
			<h2 class="section-heading" style="color: var(--sales-gold);">Inside the Library: A Preview</h2>
			<div class="content-grid" style="text-align: left;">
				<div class="content-tile">
					<span class="content-tile__number">01</span>
					<h4>Where Is the River Sidon? A Geographic Case for the Mississippi</h4>
					<p>Heartland geography examined through the text of the Book of Mormon itself.</p>
				</div>
				<div class="content-tile">
					<span class="content-tile__number">02</span>
					<h4>Haplogroup X: The DNA That Points to Ancient North America</h4>
					<p>What mitochondrial DNA research indicates about Book of Mormon populations — and why mainstream archaeology has ignored it.</p>
				</div>
				<div class="content-tile">
					<span class="content-tile__number">03</span>
					<h4>The Fortifications of Alma 50: Archaeological Parallels in the Eastern United States</h4>
					<p>Ancient fortification patterns in Ohio and the surrounding region that match the descriptions in the text.</p>
				</div>
				<div class="content-tile">
					<span class="content-tile__number">04</span>
					<h4>Joseph Smith's Own Words on Book of Mormon Geography</h4>
					<p>What the Prophet actually said — and didn't say — about the location of Book of Mormon events. Documented. Sourced. In context.</p>
				</div>
				<div class="content-tile">
					<span class="content-tile__number">05</span>
					<h4>America as the Land of Promise: The Prophetic Covenant</h4>
					<p>What the Book of Mormon, the Doctrine and Covenants, and the founding generation all point to regarding this land.</p>
				</div>
				<div class="content-tile">
					<span class="content-tile__number">06</span>
					<h4>The Signs Are Accelerating: A Scriptural Framework for Current Events</h4>
					<p>Latter-day prophecy and what the scriptures indicate about the times we're living in. No speculation — just text and timeline.</p>
				</div>
				<div class="content-tile">
					<span class="content-tile__number">07</span>
					<h4>The Stick of Joseph: Why Ezekiel 37 Matters More Than You Think</h4>
					<p>A deep reading of the prophecy, the historical context, and what it means for the Book of Mormon's role in the last days.</p>
				</div>
				<div class="content-tile">
					<span class="content-tile__number">08</span>
					<h4>Preparing Temporally: Self-Reliance Beyond 72-Hour Kits</h4>
					<p>Practical, prophetically grounded preparedness — what to actually do, and why the counsel has never been more relevant.</p>
				</div>
			</div>
		</div>
	</section>

	<!-- Section 8: Pricing — data-slot for future admin injection -->
	<section class="section section--white" data-slot="body-7">
		<div class="container text-center">
			<h2 class="section-heading" style="color: var(--sales-navy);">Choose Your Plan</h2>
			<p class="section-subheading">
				Both plans include full access to every video in the library — plus all new content as it's added.
			</p>
			<div class="pricing-grid">
				<div class="pricing-card">
					<div class="pricing-card__plan">Monthly</div>
					<div class="pricing-card__price">$9<span>.97/mo</span></div>
					<div class="pricing-card__billed">&nbsp;</div>
					<ul class="pricing-card__features">
						<li>Full library access — 1,600+ videos</li>
						<li>New content added regularly</li>
						<li>Watch on any device</li>
						<li>Cancel anytime</li>
					</ul>
					<a href="/subscription" class="btn btn--charcoal pricing-btn" on:click|preventDefault={handleSubscribe}>Subscribe Monthly</a>
				</div>
				<div class="pricing-card pricing-card--featured">
					<div class="pricing-badge">Best Value</div>
					<div class="pricing-card__plan">Annual</div>
					<div class="pricing-card__price">$7<span>.97/mo</span></div>
					<div class="pricing-card__billed">Billed annually at $95.64/year</div>
					<ul class="pricing-card__features">
						<li>Full library access — 1,600+ videos</li>
						<li>New content added regularly</li>
						<li>Watch on any device</li>
						<li>Cancel anytime</li>
						<li>Save $24 per year vs. monthly</li>
					</ul>
					<a href="/subscription" class="btn btn--gold pricing-btn" on:click|preventDefault={handleSubscribe}>Subscribe Annually</a>
				</div>
			</div>
			<p class="pricing-reassurance">Cancel anytime from your account — no questions asked.</p>
		</div>
	</section>

	<!-- Section 9: FAQ — data-slot for future admin injection -->
	<section class="section section--off-white" data-slot="body-8">
		<div class="container container--narrow">
			<h2 class="section-heading text-center" style="color: var(--sales-navy);">Frequently Asked Questions</h2>
			<div class="faq-list">
				{#each [
					{ q: "What devices can I watch on?", a: "You can stream from any device with a web browser — desktop, laptop, tablet, or phone. No special app required. Just log in and watch." },
					{ q: "How often is new content added?", a: "New videos are added regularly, including recordings from live FIRM Foundation conferences, new research presentations, and special series. The library is always growing." },
					{ q: "Is this the same content as on YouTube?", a: "Some introductory and highlight content is available on our YouTube channel, but the streaming library contains the full archive — over 1,600 videos, including complete conference sessions, deep-dive series, and content that has never been published on YouTube. This is the full collection." },
					{ q: "Can I share my account with my family?", a: "Your subscription is for individual use, but household members watching on shared devices in your home is fine. We trust your judgment." },
					{ q: "What happens if I cancel?", a: "You keep full access through the end of your current billing period. After that, your account becomes inactive. No penalty, no hassle. You can resubscribe anytime." },
					{ q: "How do I get started?", a: "Choose your plan — monthly or annual — and you'll have instant access to the full library the moment you subscribe." }
				] as faq, i}
					<div class="faq-item" class:active={openFaq === i}>
						<button class="faq-question" on:click={() => toggleFaq(i)}>
							{faq.q}
						</button>
						<div class="faq-answer">
							<p>{faq.a}</p>
						</div>
					</div>
				{/each}
			</div>
		</div>
	</section>

	<!-- Section 10: Final CTA — data-slot for future admin injection -->
	<section class="section section--navy final-cta-section" data-slot="body-9">
		<div class="container container--narrow text-center">
			<h2 class="section-heading" style="color: white;">The Evidence Is Here. It Always Has Been.</h2>
			<p class="final-cta__body">
				For over twenty years, I've been gathering, organizing, and presenting the evidence for the Book of Mormon as a real history of real people in ancient North America. Not to replace faith — but to strengthen it. Not to settle every question — but to show that the questions have answers worth studying.
			</p>
			<p class="final-cta__body">
				This library represents the work of my life and the contributions of researchers I deeply respect. It's here for you to study at your own pace, on your own time, and to share with the people in your life who are searching for solid ground.
			</p>
			<p class="final-cta__body" style="font-weight: 500; color: white;">
				The evidence shows something. Come see for yourself.
			</p>
			<div class="final-cta__signature">— Rod Meldrum</div>
			<a href="/subscription" class="btn btn--gold btn--large" on:click|preventDefault={handleSubscribe}>Subscribe Now</a>
			<p class="final-cta__fine-print">
				$9.97/mo or $7.97/mo billed annually. Cancel anytime. Full access to the entire library from day one.
			</p>
		</div>
	</section>

	<!-- Footer — sales page style -->
	<footer class="sales-footer">
		<div class="container">
			<img src="/Images/Logo.png" alt="Book of Mormon Evidence" class="footer__logo" />
			<p>&copy; 2026 Book of Mormon Evidence / FIRM Foundation. All rights reserved.</p>
			<div class="footer__links">
				
			</div>
			<p class="footer__disclaimer">
				This research and content represent independent scholarship and are not produced by, affiliated with, or endorsed by The Church of Jesus Christ of Latter-day Saints.
			</p>
		</div>
	</footer>
</div>

<style>
	:global(html) {
		scroll-behavior: smooth;
	}



	/* Card Styles TESTING*/
	.outer {
  width: 600px;
  height: 250px;
  border-radius: 10px;
  padding: 1px;
  background: radial-gradient(circle 230px at 55% 0%, rgb(255, 234, 112), rgba(68, 39, 7, 0.05));
  position: relative;
  animation: breathe 8s ease-in-out infinite;
}

.dot {
  width: 5px;
  aspect-ratio: 1;
  position: absolute;
  background-color: #ffffff;
  box-shadow: 2px 2px 5px #fff174;
  border-radius: 100px;
  z-index: 2;
  right: 10%;
  top: 10%;
  animation: moveDot 6s linear infinite;
}

@keyframes moveDot {
  0%,
  100% {
    top: 8%;
    right: 9%;
  }
  25% {
    top: 10%;
    right: calc(100% - 60px);
  }
  50% {
    top: calc(100% - 30px);
    right: calc(100% - 66px);
  }
  75% {
    top: calc(100% - 30px);
    right: 11%;
  }
}

@keyframes breathe {
	0%,
	100% {
		transform: scale(1);
	}
	50% {
		transform: scale(0.95);
	}
}

.card {
  z-index: 1;
  width: 100%;
  height: 100%;
  border-radius: 9px;
  border: solid 1px #202222;
  background-size: 20px 20px;
  background: radial-gradient(circle 280px at 0% 0%, rgba(73, 68, 59, 0.5), rgba(0, 6, 19, 0.5));
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  flex-direction: column;
  color: #e7cab2;
}
.ray {
  width: 220px;
  height: 65px;
  border-radius: 100px;
  position: absolute;
  background-color: linear-gradient(30deg, #241000 4%, #ffeb7a, #241000); 
  opacity: 0.4;
  box-shadow: 0 0 50px #fff;
  filter: blur(10px);
  transform-origin: 50%;
  top: 0%;
  left: 55;
  transform: rotate(84deg);
}

.card .text {
  font-family: 'Quicksand', serif;
  font-weight: bolder;
  font-size: 4rem;
  background: linear-gradient(45deg, #241000 4%, #ffeb7a, #241000);
  background-clip: text;
  color: transparent;
  animation: breathe 8s ease-in-out infinite;
}

.line {
  width: 100%;
  height: 1px;
  position: absolute;
  background-color: #2c2c2c;
}
.topl {
  top: 10%;
  background: linear-gradient(90deg, #888888 30%, #1d1f1f 70%);
}
.bottoml {
  bottom: 10%;
}
.leftl {
  left: 10%;
  width: 1px;
  height: 100%;
  background: linear-gradient(180deg, #747474 30%, #222424 70%);
}
.rightl {
  right: 10%;
  width: 1px;
  height: 100%;
}


	.page-wrapper {
		position: relative;
	}

	.hero-section {
		min-height: 100vh;
		width: 100%;
		position: relative;
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
	}

	.zoom-section {
		min-height: 100vh;
		width: 100%;
		position: relative;
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
	}

	.zoom-content {
		position: relative;
		width: 100%;
		height: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	/* Book Section */
	.book-section {
		background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
	}

	.book-container {
		position: relative;
		width: 100vw;
		height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		transform: scale(1);
	}

	.book-image {
		width: 100vw;
		height: 100vh;
		object-fit: cover;
		border-radius: 0;
		box-shadow: none;
		filter: drop-shadow(0 0 30px rgba(255, 215, 0, 0.3));
		filter: blur(0px);
		align-items: left;
		/* Fade in from black on page load */
		animation: fadeInFromBlack 2s ease-out forwards;
	}

	@keyframes fadeInFromBlack {
		0% {
			opacity: 0;
			filter: brightness(0) drop-shadow(0 0 30px rgba(255, 215, 0, 0.3));
		}
		100% {
			opacity: 1;
			filter: brightness(1) drop-shadow(0 0 30px rgba(255, 215, 0, 0.3));
		}
	}

	.book-glow {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: radial-gradient(circle, rgba(255, 215, 0, 0.2) 0%, transparent 70%);
		border-radius: 0;
		animation: glow-pulse 3s ease-in-out infinite;
	}

	@keyframes glow-pulse {
		0%, 100% { opacity: 0.5; transform: scale(1); }
		50% { opacity: 0.8; transform: scale(1.05); }
	}

	/* New York Section */
	.newyork-section {
		background: linear-gradient(135deg, #2c5530 0%, #3d7c47 50%, #4a9960 100%);
	}

	/* USA 1830 Section */
	.usa1830-section {
		background: linear-gradient(135deg, #8b4513 0%, #a0522d 50%, #cd853f 100%);
	}

	/* Modern USA Section */
	.usa-modern-section {
		background: linear-gradient(135deg, #1e3c72 0%, #2a5298 50%, #3b82f6 100%);
	}

	/* World Section */
	.world-section {
		background: linear-gradient(135deg, #134e5e 0%, #71b280 50%, #a8e6cf 100%);
	}

	/* Globe Section */
	.globe-section {
		background-image: url('/HOMEPAGE_TEST_ASSETS/The_Globe_World_Background.png');
		background-size: cover;
		background-position: center;
		background-repeat: no-repeat;
		position: relative;
	}

	.globe-section::before {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><circle cx="20" cy="20" r="1" fill="white" opacity="0.8"/><circle cx="80" cy="30" r="0.5" fill="white" opacity="0.6"/><circle cx="40" cy="60" r="1.5" fill="white" opacity="0.4"/><circle cx="70" cy="80" r="1" fill="white" opacity="0.7"/><circle cx="10" cy="80" r="0.8" fill="white" opacity="0.5"/></svg>');
		animation: stars-twinkle 10s linear infinite;
	}

	.globe-container {
		position: relative;
		width: 100%;
		height: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.globe {
		position: relative;
		width: 600px;
		height: 600px;
		border-radius: 50%;
		overflow: hidden;
	}

	.globe-image {
		width: 100%;
		height: 100%;
		border-radius: 50%;
		animation: globe-rotate 20s linear infinite;
	}

	.globe-atmosphere {
		position: absolute;
		top: -10px;
		left: -10px;
		right: -10px;
		bottom: -10px;
		border-radius: 50%;
		background: radial-gradient(circle, transparent 45%, rgba(135, 206, 235, 0.3) 55%, rgba(135, 206, 235, 0.1) 70%, transparent 80%);
		pointer-events: none;
	}

	@keyframes globe-rotate {
		from { transform: rotate(0deg); }
		to { transform: rotate(360deg); }
	}

	@keyframes stars-twinkle {
		0%, 100% { opacity: 0.8; }
		50% { opacity: 0.3; }
	}

	/* Map and Globe Containers */
	.map-container {
		position: relative;
		width: 100vw;
		height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.map-image {
		width: 100vw;
		height: 100vh;
		object-fit: cover;
		transition: transform 1.5s cubic-bezier(0.25, 0.46, 0.45, 0.94);
	}

	.map-image.vintage {
		filter: sepia(0.3) contrast(1.2);
	}

	/* Location Markers */
	.location-marker {
		position: absolute;
		width: 20px;
		height: 20px;
		transform: translate(-50%, -50%);
	}

	.marker-dot {
		width: 12px;
		height: 12px;
		background: #ff6b6b;
		border-radius: 50%;
		position: relative;
		z-index: 2;
		box-shadow: 0 0 10px rgba(255, 107, 107, 0.8);
	}

	.marker-pulse {
		position: absolute;
		top: -4px;
		left: -4px;
		width: 20px;
		height: 20px;
		border: 2px solid #ff6b6b;
		border-radius: 50%;
		animation: marker-pulse 2s ease-out infinite;
	}

	@keyframes marker-pulse {
		0% { transform: scale(0.5); opacity: 1; }
		100% { transform: scale(2); opacity: 0; }
	}

	/* Marker Positions */
	.hill-cumorah { top: 35%; left: 85%; }
	.palmyra { top: 25%; left: 75%; }
	.kirtland { top: 30%; left: 70%; }
	.utah { top: 45%; left: 25%; }
	.missouri { top: 50%; left: 55%; }
	.americas { top: 60%; left: 30%; }
	.middle-east { top: 40%; left: 75%; }

	/* Content Overlays */
	.content-overlay {
		position: absolute;
		z-index: 10;
		color: white;
		text-align: center;
		max-width: 800px;
		padding: 2.5rem;
		background: rgba(0, 0, 0, 0.4);
		backdrop-filter: blur(15px);
		border-radius: 20px;
		border: 1px solid rgba(255, 255, 255, 0.2);
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
	}

	.book-section .content-overlay {
		top: 10%;
		left: 50%;
		transform: translateX(-50%);
		background: rgba(0, 0, 0, 0.5);
		backdrop-filter: blur(20px);
	}

	.newyork-section .content-overlay,
	.usa1830-section .content-overlay,
	.usa-modern-section .content-overlay,
	.world-section .content-overlay {
		bottom: 10%;
		left: 50%;
		transform: translateX(-50%);
	}

	.globe-section .content-overlay {
		top: 15%;
		left: 50%;
		transform: translateX(-50%);
		background: rgba(0, 20, 40, 0.6);
		backdrop-filter: blur(20px);
		border: 1px solid rgba(135, 206, 235, 0.3);
	}

	/* Typography */
	.main-title {
		font-size: 4rem;
		font-weight: 800;
		margin-bottom: 1.5rem;
		text-shadow: 3px 3px 6px rgba(0, 0, 0, 0.8);
		filter: drop-shadow(0 0 10px rgba(255, 215, 0, 0.3));
	}

	.title-line {
		display: block;
		background: linear-gradient(135deg, #ffd700 0%, #ffed4e 100%);
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
		text-shadow: none;
	}

	.section-title {
		font-size: 3rem;
		font-weight: 700;
		margin-bottom: 1.5rem;
		text-shadow: 3px 3px 6px rgba(0, 0, 0, 0.9);
		color: #ffffff;
		filter: drop-shadow(0 0 8px rgba(255, 255, 255, 0.2));
	}

	.subtitle,
	.section-description {
		font-size: 1.5rem;
		margin-bottom: 2rem;
		opacity: 1;
		color: rgba(255, 255, 255, 0.95);
		text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
		line-height: 1.6;
	}

	/* Buttons */
	.cta-button,
	.explore-button {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 1rem 2rem;
		background: linear-gradient(135deg, #ff6b6b 0%, #ee5a24 100%);
		color: white;
		border: none;
		border-radius: 50px;
		font-size: 1.1rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.3s ease;
		box-shadow: 0 10px 30px rgba(255, 107, 107, 0.4);
	}

	.cta-button:hover,
	.explore-button:hover {
		transform: translateY(-3px);
		box-shadow: 0 15px 40px rgba(255, 107, 107, 0.6);
	}

	.join-button {
		display: inline-block;
		padding: 1rem 2rem;
		background: transparent;
		color: white;
		border: 2px solid white;
		border-radius: 50px;
		text-decoration: none;
		font-weight: 600;
		margin-left: 1rem;
		transition: all 0.3s ease;
	}

	.join-button:hover {
		background: white;
		color: #004e92;
	}

	/* Info Cards and Stats */
	.info-card {
		background: rgba(255, 255, 255, 0.15);
		backdrop-filter: blur(20px);
		padding: 2rem;
		border-radius: 20px;
		margin-top: 2rem;
		border: 1px solid rgba(255, 255, 255, 0.3);
		box-shadow: 0 15px 35px rgba(0, 0, 0, 0.2);
	}

	.info-card h3 {
		color: #ffd700;
		font-weight: 700;
		font-size: 1.3rem;
		margin-bottom: 1rem;
		text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.7);
	}

	.info-card p {
		color: rgba(255, 255, 255, 0.95);
		line-height: 1.6;
		text-shadow: 1px 1px 3px rgba(0, 0, 0, 0.8);
	}

	.timeline-info {
		display: flex;
		gap: 2rem;
		justify-content: center;
		margin-top: 2rem;
	}

	.timeline-item {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		text-align: center;
		background: rgba(255, 255, 255, 0.1);
		backdrop-filter: blur(15px);
		padding: 1.5rem;
		border-radius: 15px;
		border: 1px solid rgba(255, 255, 255, 0.2);
		box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
		min-width: 150px;
	}

	.year {
		font-size: 1.5rem;
		font-weight: 700;
		color: #ffd700;
		text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
		filter: drop-shadow(0 0 5px rgba(255, 215, 0, 0.4));
	}

	.event {
		font-size: 1rem;
		opacity: 1;
		color: rgba(255, 255, 255, 0.95);
		text-shadow: 1px 1px 3px rgba(0, 0, 0, 0.8);
		line-height: 1.4;
	}

	.stats-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 2rem;
		margin-top: 2rem;
	}

	.stat-item {
		text-align: center;
		background: rgba(255, 255, 255, 0.1);
		backdrop-filter: blur(15px);
		padding: 2rem 1.5rem;
		border-radius: 20px;
		border: 1px solid rgba(255, 255, 255, 0.2);
		box-shadow: 0 15px 30px rgba(0, 0, 0, 0.2);
		transition: transform 0.3s ease, box-shadow 0.3s ease;
	}

	.stat-item:hover {
		transform: translateY(-5px);
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
	}

	.stat-number {
		display: block;
		font-size: 2.5rem;
		font-weight: 800;
		color: #ffd700;
		text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
		filter: drop-shadow(0 0 8px rgba(255, 215, 0, 0.4));
		margin-bottom: 0.5rem;
	}

	.stat-label {
		font-size: 1rem;
		opacity: 1;
		color: rgba(255, 255, 255, 0.95);
		text-shadow: 1px 1px 3px rgba(0, 0, 0, 0.8);
		line-height: 1.4;
	}

	/* Hero Title and Quote - Base Styles */
	.main-title-container {
		position: absolute;
		top: 15%;
		left: 50%;
		transform: translateX(-50%);
		text-align: center;
		z-index: 15;
		max-width: clamp(1200px, 90vw, 2000px);
		background: rgba(14, 18, 70, 0.75);
		border-radius: 25px;
		padding: clamp(1rem, 2vw, 1rem) clamp(2rem, 5vw, 5rem) 0 clamp(2rem, 5vw, 5rem);
	}

	.hero-title {
		font-family: 'Garamond', Times, serif;
		font-weight: 900;
		color: #ffd700;
		text-shadow: 4px 4px 8px rgba(0, 0, 0, 0.9);
		filter: drop-shadow(0 0 15px rgba(255, 215, 0, 0.4));
		margin-bottom: 2rem;
		line-height: 0.75;
		padding: 1rem 1.5rem;
	}

	.hero-title-book {
		font-size: clamp(3rem, 5vw, 4.5rem);
	}

	.hero-title-evidence {
		font-size: clamp(8rem, 15vw, 13.5rem);
	}

	.hero-title-beta {
		font-size: clamp(4rem, 7vw, 6.5rem);
		background: linear-gradient(45deg,#8a7658 , #ffed4e , #ffffff, #ffed4e , #daa30cdd , #7a5926	);
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
		text-shadow: none;
		transition: all 0.5s ease;
		animation: grow-text 2s ease;
	}

	@keyframes grow-text {
		0% {
			font-size: 0rem;
		}
		100% {
			font-size: clamp(4rem, 7vw, 6.5rem);
		}
	}

	.hero-quote {
		font-family: 'Quicksand_';
		font-size: clamp(1.1rem, 1.5vw, 1.4rem);
		color: rgba(255, 255, 255, 0.95);
		text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
		font-style: italic;
		line-height: 1.6;
		margin-bottom: 3rem;
		max-width: 800px;
		/*margin-left: auto;
		margin-right: auto;*/
	}

	.hero-quote p {
		color: white;
		font-family: 'Ayres';
		font-size: clamp(1.1rem, 5vw, 3rem);
	}

	.hero-quote cite {
		display: block;
		margin-top: 1rem;
		font-size: clamp(1.1rem, 5vw, 2.5rem);
		color: #ffd700;
		font-style: normal;
		font-weight: 600;
		font-family: 'Ayres';
	}

	/* Sales-page design tokens (from streaming-sales-page) */
	.page-wrapper.home-page {
		--sales-navy: #0D1B3E;
		--sales-gold: #B7953B;
		--sales-gold-light: #D4B255;
		--sales-charcoal: #1E2022;
		--sales-white: #FFFFFF;
		--sales-off-white: #F7F8FA;
		--sales-light-gray: #E8EAF0;
		--sales-text-muted: #9CA3AF;
		--sales-heading-font: 'Playfair Display', Georgia, serif;
		--sales-body-font: 'Poppins', 'Roboto', sans-serif;
	}

	/* Sales page sections */
	.section {
		padding: 80px 0;
	}
	.section--navy {
		background: var(--sales-navy);
		color: var(--sales-white);
	}
	.section--white {
		background: var(--sales-white);
		color: var(--sales-charcoal);
	}
	.section--off-white {
		background: var(--sales-off-white);
		color: var(--sales-charcoal);
	}

	.container {
		position: relative;
		max-width: 100vw;
		margin: 0 auto;
		padding: 0 24px;
	}

	.container--narrow {
		max-width: 780px;
		margin: 0 auto;
		padding: 0 24px;
	}

	.section-heading {
		font-family: var(--sales-heading-font);
		font-size: 38px;
		font-weight: 600;
		line-height: 1.25;
		margin-bottom: 20px;
	}
	.section-subheading {
		font-family: var(--sales-body-font);
		font-size: 17px;
		font-weight: 300;
		line-height: 1.8;
		max-width: 700px;
		margin: 0 auto;
	}

	.tension p {
		font-size: 17px;
		line-height: 1.9;
		margin-bottom: 20px;
		color: var(--sales-charcoal);
	}
	.tension p:last-child {
		font-weight: 600;
		font-style: italic;
		color: var(--sales-navy);
		margin-bottom: 0;
	}

	.container-grove {
		background: linear-gradient(to right, rgb(255, 255, 255, 0) 0%, rgb(247, 248, 250, 1) 30%, rgb(247, 248, 250, 1) 70%, rgba(255, 255, 255, 0) 100%);
		position: relative;
		z-index: 1;
		width: 100%;
		height: 100%;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
	}

	.image-grove-left {
		position: absolute;
		left: 0;
		top: 0;
		width: auto;
		height: 100%;
		z-index: 0;
	}
	.image-grove-right {
		position: absolute;
		right: 0;
		top: 0;
		width: auto;
		height: 100%;
		
		z-index: 0;
	}

	.categories-grid {
		position: relative;
		display: flex;
		flex-direction: row;
		flex-wrap: wrap;
		gap: 24px;
		margin-top: 48px;
		margin: auto 0;
		width: 100%;
		justify-content: center;
		z-index: 1;
	}
	.category-card {
		background: var(--sales-white);
		border: 1px solid var(--sales-light-gray);
		border-left: 4px solid var(--sales-gold);
		padding: 32px 28px;
		transition: box-shadow 0.3s ease;
		width: 250px;
	}
	.category-card:hover {
		box-shadow: 5px 5px 20px rgba(0,0,0,0.08);
	}
	.category-card h3 {
		font-size: 20px;
		color: var(--sales-navy);
		margin-bottom: 12px;
	}
	.category-card p {
		font-size: 15px;
		line-height: 1.75;
		color: #555;
	}
	.categories-grid .category-card:nth-child(5) {
		grid-column: 1 / -1;
		max-width: 520px;
		justify-self: center;
	}
	.library-closing {
		text-align: center;
		margin-top: 40px;
		font-size: 15px;
		font-style: italic;
		color: var(--sales-text-muted);
	}

	.stats-banner {
		padding: 60px 0;
	}
	.sales-stats-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 20px;
		text-align: center;
	}
	.sales-stat-item .stat-item__number {
		font-family: var(--sales-heading-font);
		font-size: 48px;
		font-weight: 700;
		color: var(--sales-gold);
		line-height: 1;
		margin-bottom: 8px;
	}
	.sales-stat-item .stat-item__label {
		font-size: 14px;
		text-transform: uppercase;
		letter-spacing: 1.5px;
		color: rgba(255,255,255,0.7);
		font-weight: 400;
	}

	.speakers-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 32px;
		margin-top: 48px;
		text-align: center;
	}
	.speaker-card {
		padding: 24px 16px;
	}
	.speaker-card__icon {
		width: 80px;
		height: 80px;
		border-radius: 50%;
		background: var(--sales-navy);
		margin: 0 auto 20px;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.speaker-card__icon span {
		font-family: var(--sales-heading-font);
		font-size: 28px;
		font-weight: 700;
		color: var(--sales-gold);
	}
	.speaker-card h4 {
		font-size: 18px;
		color: var(--sales-navy);
		margin-bottom: 8px;
	}
	.speaker-card p {
		font-size: 14px;
		line-height: 1.6;
		color: #666;
	}

	.for-you-list {
		list-style: none;
		margin-top: 40px;
	}
	.for-you-list li {
		position: relative;
		padding: 16px 0 16px 36px;
		font-size: 16px;
		line-height: 1.8;
		border-bottom: 1px solid var(--sales-light-gray);
	}
	.for-you-list li:last-child {
		border-bottom: none;
	}
	.for-you-list li::before {
		content: '';
		position: absolute;
		left: 0;
		top: 24px;
		width: 12px;
		height: 12px;
		border: 2px solid var(--sales-gold);
		border-radius: 2px;
		background: transparent;
	}
	.for-you-list li::after {
		content: '';
		position: absolute;
		left: 3px;
		top: 28px;
		width: 6px;
		height: 3px;
		border-left: 2px solid var(--sales-gold);
		border-bottom: 2px solid var(--sales-gold);
		transform: rotate(-45deg);
	}

	.content-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 20px;
		margin-top: 48px;
	}
	.content-tile {
		background: rgba(255,255,255,0.06);
		border: 1px solid rgba(255,255,255,0.1);
		padding: 28px 24px;
		transition: all 0.3s ease;
	}
	.content-tile:hover {
		background: rgba(255,255,255,0.1);
		border-color: var(--sales-gold);
	}
	.content-tile__number {
		font-family: var(--sales-heading-font);
		font-size: 14px;
		color: var(--sales-gold);
		margin-bottom: 8px;
		display: block;
	}
	.content-tile h4 {
		font-size: 17px;
		font-weight: 600;
		font-family: var(--sales-body-font);
		color: var(--sales-white);
		margin-bottom: 8px;
		line-height: 1.4;
	}
	.content-tile p {
		font-size: 14px;
		line-height: 1.65;
		color: rgba(255,255,255,0.6);
	}

	.pricing-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 32px;
		margin-top: 48px;
		max-width: 800px;
		margin-left: auto;
		margin-right: auto;
	}
	.pricing-card {
		border: 1px solid var(--sales-light-gray);
		padding: 40px 32px;
		text-align: center;
		position: relative;
		background: var(--sales-white);
		transition: box-shadow 0.3s ease;
	}
	.pricing-card:hover {
		box-shadow: 5px 5px 20px rgba(0,0,0,0.08);
	}
	.pricing-card--featured {
		border: 2px solid var(--sales-gold);
	}
	.pricing-badge {
		position: absolute;
		top: -14px;
		left: 50%;
		transform: translateX(-50%);
		background: var(--sales-gold);
		color: var(--sales-navy);
		font-size: 12px;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 1px;
		padding: 6px 20px;
	}
	.pricing-card__plan {
		font-family: var(--sales-heading-font);
		font-size: 20px;
		color: var(--sales-navy);
		margin-bottom: 16px;
	}
	.pricing-card__price {
		font-size: 48px;
		font-weight: 700;
		color: var(--sales-navy);
		line-height: 1;
	}
	.pricing-card__price span {
		font-size: 18px;
		font-weight: 400;
		color: var(--sales-text-muted);
	}
	.pricing-card__billed {
		font-size: 13px;
		color: var(--sales-text-muted);
		margin-top: 4px;
		margin-bottom: 28px;
		min-height: 20px;
	}
	.pricing-card__features {
		list-style: none;
		text-align: left;
		margin-bottom: 32px;
	}
	.pricing-card__features li {
		padding: 8px 0;
		font-size: 14px;
		color: #555;
		border-bottom: 1px solid #f0f0f0;
		padding-left: 24px;
		position: relative;
	}
	.pricing-card__features li::before {
		content: "\2713";
		position: absolute;
		left: 0;
		color: var(--sales-gold);
		font-weight: 700;
	}
	.pricing-card__features li:last-child {
		border-bottom: none;
	}
	.pricing-btn {
		width: 100%;
	}
	.pricing-reassurance {
		text-align: center;
		margin-top: 32px;
		font-size: 14px;
		color: var(--sales-text-muted);
	}

	.btn--charcoal {
		background: var(--sales-charcoal);
		color: var(--sales-white);
		border-color: var(--sales-charcoal);
	}
	.btn--charcoal:hover {
		background: transparent;
		color: var(--sales-charcoal);
		border-color: var(--sales-charcoal);
	}

	.faq-list {
		margin-top: 48px;
	}
	.faq-item {
		border-bottom: 1px solid var(--sales-light-gray);
		padding: 24px 0;
	}
	.faq-item:first-child {
		border-top: 1px solid var(--sales-light-gray);
	}
	.faq-question {
		font-family: var(--sales-body-font);
		font-weight: 600;
		font-size: 16px;
		color: var(--sales-navy);
		cursor: pointer;
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 16px;
		width: 100%;
		text-align: left;
		background: none;
		border: none;
	}
	.faq-question::after {
		content: "+";
		font-size: 22px;
		font-weight: 300;
		color: var(--sales-gold);
		flex-shrink: 0;
		transition: transform 0.3s ease;
	}
	.faq-item.active .faq-question::after {
		content: "\2212";
	}
	.faq-answer {
		max-height: 0;
		overflow: hidden;
		transition: max-height 0.4s ease, padding 0.4s ease;
	}
	.faq-item.active .faq-answer {
		max-height: 300px;
		padding-top: 16px;
	}
	.faq-answer p {
		font-size: 15px;
		line-height: 1.8;
		color: #555;
	}

	.final-cta-section {
		padding: 100px 0 !important;
	}
	.final-cta__body {
		font-size: 17px;
		line-height: 1.9;
		color: rgba(255,255,255,0.85);
		max-width: 700px;
		margin: 0 auto 12px;
	}
	.final-cta__signature {
		font-family: var(--sales-heading-font);
		font-size: 20px;
		font-style: italic;
		color: var(--sales-gold);
		margin-top: 24px;
		margin-bottom: 40px;
	}
	.final-cta__fine-print {
		margin-top: 16px;
		font-size: 13px;
		color: rgba(255,255,255,0.45);
	}

	.sales-footer {
		background: #080F21;
		color: rgba(255,255,255,0.5);
		padding: 40px 0;
		text-align: center;
		font-size: 13px;
		line-height: 1.8;
	}
	.sales-footer .footer__logo {
		width: 140px;
		margin: 0 auto 16px;
		opacity: 0.6;
	}
	.sales-footer .footer__links {
		margin-top: 12px;
	}
	.sales-footer .footer__links a {
		color: rgba(255,255,255,0.5);
		margin: 0 12px;
		font-size: 13px;
		transition: color 0.3s ease;
	}
	.sales-footer .footer__links a:hover {
		color: var(--sales-gold);
	}
	.sales-footer .footer__disclaimer {
		margin-top: 20px;
		font-style: italic;
		font-size: 12px;
		color: rgba(255,255,255,0.35);
		max-width: 600px;
		margin-left: auto;
		margin-right: auto;
	}

	/* Hero content slot — sales copy + Subscribe CTA */
	.hero-content-slot {
		flex-direction: column;
		justify-content: flex-end;
		align-items: center;
		gap: clamp(1rem, 2vw, 1.5rem);
		text-align: center;
		bottom: 8%;
		padding: 0 2rem;
	}

	.hero-logo {
		width: clamp(140px, 12vw, 220px);
		margin-bottom: clamp(1rem, 2vw, 2rem);
		opacity: 0.95;
	}

	.hero-copy {
		max-width: 850px;
		padding: clamp(1.5rem, 3vw, 2.5rem);
		background: rgba(13, 27, 62, 0.85);
		border-radius: 16px;
		backdrop-filter: blur(12px);
		border: 1px solid rgba(255, 255, 255, 0.12);
		box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
		opacity: 0;
		animation: hero-copy-fade-in 2.6s ease-out 5s forwards;
	}

	@keyframes hero-copy-fade-in {
		to { opacity: 1; }
	}

	.hero-headline {
		font-family: var(--sales-heading-font);
		font-size: clamp(2rem, 4.5vw, 3rem);
		font-weight: 700;
		line-height: 1.15;
		color: white;
		text-shadow: 0 2px 8px rgba(0, 0, 0, 0.6);
		margin: 0 0 1rem;
	}

	.hero-sub {
		font-family: var(--sales-body-font);
		font-size: clamp(1rem, 1.3vw, 1.125rem);
		font-weight: 300;
		line-height: 1.8;
		color: rgba(255, 255, 255, 0.95);
		max-width: 680px;
		margin: 0 auto;
		text-shadow: 0 1px 4px rgba(0, 0, 0, 0.5);
	}

	.hero-pricing {
		font-family: var(--sales-body-font);
		font-size: clamp(0.8rem, 1vw, 0.875rem);
		color: rgba(255, 255, 255, 0.85);
		font-weight: 300;
		margin: 0.5rem 0 0;
		text-shadow: 0 1px 3px rgba(0, 0, 0, 0.5);
	}

	/* Sales page button styles (from streaming-sales-page.html) */
	.btn {
		display: inline-block;
		font-family: var(--sales-body-font);
		font-weight: 600;
		font-size: 14px;
		letter-spacing: 1px;
		text-transform: uppercase;
		padding: 16px 40px;
		border: 2px solid transparent;
		cursor: pointer;
		transition: all 0.3s ease;
		text-decoration: none;
		border-radius: 4px;
	}

	.btn--gold {
		background: var(--sales-gold);
		color: var(--sales-navy);
		border-color: var(--sales-gold);
	}

	.btn--gold:hover {
		background: transparent;
		color: var(--sales-gold);
		border-color: var(--sales-gold);
	}

	.btn--large {
		padding: 18px 48px;
		font-size: 15px;
	}

	.subscribe-btn {
		align-self: center;
	}

	/* Navigation Cards - Base Styles (legacy nav-card layout) */
	.navigation-cards {
		width: 100vw;
		position: absolute;
		bottom: 11%;
		left: 50%;
		transform: translateX(-50%);
		display: flex;
		justify-content: space-between;
		gap: clamp(3rem, 15vw, 7rem);
		padding: 0 8rem;
		z-index: 15;
	}

	.nav-card {
		background: rgba(59, 138, 241, 0.05);
		backdrop-filter: blur(10px);
		border-radius: 15px;
		padding: 0;
		text-align: center;
		color: white;
		text-decoration: none;
		transition: all 0.3s ease;
		border: 1px solid rgba(255, 255, 255, 0.2);
		box-shadow: 10px 10px 20px rgba(0, 0, 0, 0.2),
					-10px -10px 20px rgba(13, 3, 41, 0.5),
					inset 0 0 10px rgba(155, 112, 63, 0.5),
					inset 0 0 10px rgba(0, 0, 0, 0.2);
		width: clamp(300px, 25vw, 400px);
		min-height: clamp(250px, 20vw, 300px);
		display: flex;
		flex-direction: column;
		justify-content: flex-start;
		align-items: stretch;
		overflow: hidden;
	}

	.nav-card:hover {
		transform: translateY(-5px);
		box-shadow: 15px 15px 30px rgba(0, 0, 0, 0.3),
					-15px -15px 30px rgba(13, 3, 41, 0.6),
					inset 0 0 15px rgba(255, 255, 255, 0.6),
					inset 0 0 15px rgba(0, 0, 0, 0.3);
		background: rgba(49, 49, 49, 0.7);
	}

	.nav-card:hover .card-action {
		background: rgba(66, 66, 66, 0.2);
	}

	.card-action {
		font-size: clamp(2rem, 3vw, 3.1rem);
		font-weight: 900;
		color: #ffd700;
		text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.9);
		filter: drop-shadow(0 0 8px rgba(255, 215, 0, 0.5));
		letter-spacing: 2px;
		text-transform: uppercase;
		width: 100%;
		height: 20%;
		display: flex;
		align-items: center;
		justify-content: center;
		margin: 0;
		background: rgba(255, 215, 0, 0.1);
		border-radius: 15px 15px 0 0;
		border-bottom: 2px solid rgba(255, 215, 0, 0.3);
		flex-shrink: 0;
	}

	.nav-card-content {
		flex: 1;
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
		padding: 1.5rem;
	}

	.card-icon {
		width: 60px;
		height: 60px;
		margin-bottom: 1rem;
		color: #ffd700;
		filter: drop-shadow(0 0 8px rgba(255, 215, 0, 0.4));
	}

	.nav-card h3 {
		font-size: 1.3rem;
		font-weight: 700;
		margin-bottom: 1rem;
		color: white;
		text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
	}

	.nav-card p {
		font-size: 1rem;
		color: rgba(255, 255, 255, 0.9);
		text-shadow: 1px 1px 3px rgba(0, 0, 0, 0.8);
		line-height: 1.4;
		margin: 0;
	}

	/* Navigation Dots */
	.scroll-navigation {
		position: fixed;
		right: 2rem;
		top: 50%;
		transform: translateY(-50%);
		display: flex;
		flex-direction: column;
		gap: 1rem;
		z-index: 100;
		background: rgba(0, 0, 0, 0.2);
		backdrop-filter: blur(10px);
		padding: 1rem 0.5rem;
		border-radius: 25px;
		border: 1px solid rgba(255, 255, 255, 0.2);
	}

	.nav-dot {
		width: 12px;
		height: 12px;
		border: 2px solid rgba(255, 255, 255, 0.7);
		border-radius: 50%;
		background: transparent;
		cursor: pointer;
		transition: all 0.3s ease;
		padding: 0;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
	}

	.nav-dot.active {
		background: white;
		border-color: white;
		transform: scale(1.2);
		box-shadow: 0 0 15px rgba(255, 255, 255, 0.5);
	}

	.nav-dot:hover {
		border-color: white;
		transform: scale(1.1);
		background: rgba(255, 255, 255, 0.3);
	}

	/* Scroll Indicator */
	.scroll-indicator {
		position: fixed;
		bottom: 2rem;
		left: 50%;
		transform: translateX(-50%);
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.5rem;
		color: white;
		z-index: 100;
		transition: opacity 0.3s ease;
		/*background: rgba(0, 0, 0, 0.3);
		backdrop-filter: blur(10px);
		padding: 1rem 1.5rem;*/
		border-radius: 25px;
		border: 1px solid rgba(255, 255, 255, 0.2);
	}

	.scroll-indicator.hidden {
		opacity: 0;
		pointer-events: none;
	}

	.scroll-text {
		font-size: 0.9rem;
		opacity: 1;
		color: rgba(255, 255, 255, 0.95);
		text-shadow: 1px 1px 3px rgba(0, 0, 0, 0.8);
		font-weight: 500;
	}

	.scroll-arrow {
		width: 24px;
		height: 24px;
		animation: bounce 2s ease-in-out infinite;
		filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.5));
	}

	@keyframes bounce {
		0%, 20%, 50%, 80%, 100% { transform: translateY(0); }
		40% { transform: translateY(-10px); }
		60% { transform: translateY(-5px); }
	}

	/* Zoom Animations */
	.zoom-section {
		transform: scale(1.1);
		opacity: 0.7;
	}

	.zoom-section.active {
		transform: scale(1);
		opacity: 1;
	}

	.zoom-section.past {
		transform: scale(0.9);
		opacity: 0.3;
	}

	.zoom-section.future {
		transform: scale(1.2);
		opacity: 0.5;
	}
	/* Fixed Header for Slide 1 */
	.home_page_fixed_header {
		position: fixed;
		top: 0;
		left: 50%;
		transform: translateX(-50%);
		padding: 2rem 0;
		text-align: center;
		color: white;
		width: 100vw;
		height: 6vh;
		background: var(--bg-ghost);
		border-radius: 0 0 15px 15px;
		/*box-shadow: 10px 10px 20px rgba(0, 0, 0, 0.2),
					-10px -10px 20px rgba(13, 3, 41, 0.5),
					inset 0 0 10px rgba(255, 255, 255, 0.5),
					inset 0 0 10px rgba(0, 0, 0, 0.2);*/
		backdrop-filter: blur(10px);
		pointer-events: auto;
		z-index: 50;
	}
	/* Responsive Design */
	
	/* High Resolution Displays (2K+) - 1440p and above */
	@media (min-height: 1440px) {
		.main-title-container {
			padding: 2rem 6rem 0 6rem;
			max-width: 2200px;
		}
		
		.hero-title {
			padding: 1.5rem 2rem;
			margin-bottom: 3rem;
		}
		
		.navigation-cards {
			gap: 6rem;
			bottom: 8%;
		}
		
		.nav-card {
			width: 450px;
			min-height: 350px;
		}
	}

	/* Standard 1080p Displays */
	@media (max-height: 1080px) and (min-width: 1024px) {
		.main-title-container {
			padding: 1rem 4rem 0 4rem;
			max-width: 1800px;
			top: 10%;
		}
		
		.hero-title {
			padding: 0.8rem 1.2rem;
			margin-bottom: 1.5rem;
			
		}
		
		.hero-title-book {
			font-size: clamp(2.5rem, 4vw, 3.5rem);
		}

		.hero-title-evidence {
			font-size: clamp(6rem, 12vw, 10rem);
		}

		.hero-title-beta {
			font-size: clamp(3rem, 5.5vw, 5rem);
		}
		
		.hero-quote p {
			font-size: 2.5rem !important;
		}
		
		.navigation-cards {
			gap: 5rem;
			bottom: 5%;
			left: 50%;
			width: 100vw;
			box-sizing: border-box;
			height: 87vh;
			padding: 0 0;
		}
		
		.nav-card {
			width: 550px;
			min-height: 80px;
			padding: 0.5rem;
		}

		.nav-card-content {
			padding: 0.25rem
		}
		
		.card-action {
			font-size: 2.5rem;
		}
		
		.nav-card h3 {
			font-size: clamp(1rem, 1.5vw, 1.1rem);
		}
		
		.nav-card p {
			font-size: 0.9rem;
		}

		.watch-card {
			display: flex;
			justify-content: center;
			align-items: end;
			width: 100vw;
			height: 100%;
			box-sizing: border-box;
			max-width: 600px;
		}

		.watch-card .outer {
			width: 80%;
			height:35%;
		}

		.watch-card .outer .card {
			height: 100%;
		}

		.watch-card .outer .card .text {
			font-size: 3.25rem;
		}

	}

	@media (max-width: 1024px) {
		.hero-quote p {
			font-size: 2.5rem !important;
		}
		
		.navigation-cards {
			gap: 5rem;
			bottom: 5%;
			left: 50%;
			width: 100vw;
			box-sizing: border-box;
			height: 87vh;
			padding: 0 0;
		}
		
		.nav-card {
			width: 550px;
			min-height: 80px;
			padding: 0.5rem;
		}

		.nav-card-content {
			padding: 0.25rem
		}
		
		.card-action {
			font-size: 2.5rem;
		}
		
		.nav-card h3 {
			font-size: clamp(1rem, 1.5vw, 1.1rem);
		}
		
		.nav-card p {
			font-size: 0.9rem;
		}

		.watch-card {
			display: flex;
			justify-content: center;
			align-items: end;
			width: 100vw;
			height: 100%;
			box-sizing: border-box;
			
		}

		.watch-card .outer {
			width: 80%;
			height:35%;
		}

		.watch-card .outer .card {
			height: 100%;
		}

		.watch-card .outer .card .text {
			font-size: 3rem;
		}
	}

	/* Sales sections responsive */
	@media (max-width: 768px) {
		.section-heading {
			font-size: 28px;
		}
		.sales-stats-grid {
			grid-template-columns: repeat(2, 1fr);
			gap: 32px;
		}
		.sales-stat-item .stat-item__number {
			font-size: 36px;
		}
		.content-grid {
			grid-template-columns: 1fr;
		}
		.pricing-grid {
			grid-template-columns: 1fr;
			max-width: 400px;
		}
		.speakers-grid {
			grid-template-columns: repeat(2, 1fr);
			gap: 20px;
		}
		.categories-grid {
			grid-template-columns: 1fr;
		}
		.categories-grid .category-card:nth-child(5) {
			max-width: 100%;
		}
		.btn {
			padding: 14px 32px;
			font-size: 13px;
		}
		.btn--large {
			padding: 16px 36px;
		}
	}

	@media (max-width: 480px) {
		.speakers-grid {
			grid-template-columns: 1fr;
		}
		.sales-stats-grid {
			grid-template-columns: repeat(2, 1fr);
		}
	}

	@media (max-width: 768px) {
		.main-title-container {
			top: 5%;
		}
		.main-title {
			font-size: 2.5rem;
		}

		.section-title {
			font-size: 2rem;
		}

		.subtitle,
		.section-description {
			font-size: 1.2rem;
		}

		.content-overlay {
			padding: 2rem 1.5rem;
			max-width: 90%;
			margin: 0 auto;
		}

		.hero-title {
			font-size: 3rem;
		}

		.hero-headline {
			font-size: clamp(1.75rem, 4vw, 2.5rem);
		}

		.hero-sub {
			font-size: 1rem;
		}

		.navigation-cards {
			flex-direction: column;
			gap: 1.5rem;
			align-items: center;
			bottom: 8%;
			height: 87vh;
			width: 100vw;
			padding: 0 0;
		}

		.nav-card {
			width: 250px;
			min-height: 160px;
			padding: 0;
		}

		.nav-card-content {
			padding: 1rem;
		}

		.card-icon {
			width: 50px;
			height: 50px;
			margin-bottom: 1rem;
		}

		.card-action {
			font-size: 1.8rem;
		}

		.nav-card h3 {
			font-size: 1.2rem;
		}

		.nav-card p {
			font-size: 0.9rem;
		}

		.watch-card {
			display: flex;
			justify-content: center;
			align-items: end;
			width: 100vw;
			height: 100%;
			box-sizing: border-box;
			
		}

		.watch-card .outer {
			width: 80%;
			height:45%;
		}

		.watch-card .outer .card {
			height: 100%;
		}

		.watch-card .outer .card .text {
			font-size: 2.5rem;
		}


		.globe {
			width: 400px;
			height: 400px;
		}

		.timeline-item {
			min-width: 200px;
			max-width: 250px;
		}

		.stats-grid {
			grid-template-columns: 1fr;
			gap: 1rem;
			max-width: 300px;
			margin: 2rem auto 0;
		}

		.scroll-navigation {
			right: 1rem;
			padding: 0.8rem 0.4rem;
		}

		.scroll-indicator {
			padding: 0.8rem 1.2rem;
		}

		.final-cta {
			display: flex;
			flex-direction: column;
			gap: 1rem;
			align-items: center;
		}

		.join-button {
			margin-left: 0;
		}
	}

	@media (max-width: 480px) {
		/* Universal box-sizing fix */
		*,
		*::before,
		*::after {
			box-sizing: border-box;
		}

		.scroll-container {
			height: 100vh;
			overflow: hidden;
			position: relative;
			width: 100vw;
			margin: 0;
			padding: 0;
			/* Enable touch scrolling on mobile */
			-webkit-overflow-scrolling: touch;
			touch-action: pan-y;
		}

		.main-title {
			font-size: 2rem;
		}

		.section-title {
			font-size: 1.5rem;
		}

		.content-overlay {
			padding: 0.25rem ;
			max-width: 100vw;
			margin: 0.75rem 0;
			box-sizing: border-box;
			min-width: 95vw;
			bottom: 50%;
			left: 0;
			transform: translateX(0%);
			height: max-content;
		}

		/* Main title container mobile fixes */
		.main-title-container {
			padding: 0;
			width: 100vw;
			margin: 1rem 0;
			box-sizing: border-box;
			border-radius: 0;
			top: 6%;
		}

		.hero-title {
			font-size: 2.2rem;
			padding: 0.5rem 1rem;
		}

		.hero-title-book {
			font-size: 1.5rem;
		}

		.hero-title-evidence {
			font-size: clamp(2.75rem, 16vw, 6rem);
		}

		@keyframes grow-text {
		0% {
			font-size: 0rem;
		}
		100% {
			font-size: 3rem;
		}
	}

		.hero-title-beta {
			font-size: 3rem;
		}

		.hero-headline {
			font-size: clamp(1.5rem, 6vw, 2rem);
		}

		.hero-sub {
			font-size: 0.95rem;
		}

		.hero-pricing {
			font-size: 0.8rem;
		}

		/* Navigation cards mobile optimization */
		.navigation-cards {
			bottom: 10%;
			gap: 1rem;
			padding: 0 1rem;
			width: calc(100vw - 2rem);
			box-sizing: border-box;
			height: 87vh;
		}

		.nav-card {
			width: calc(100vw - 3rem);
			max-width: 280px;
			min-height: 120px;
			padding: 0;
			margin: 0 auto;
			box-sizing: border-box;
		}

		.nav-card-content {
			padding: 0.75rem;
		}

		.card-icon {
			width: 35px;
			height: 35px;
		}

		.card-action {
			font-size: 1.5rem;
			height: 15%;
		}

		.nav-card h3 {
			font-size: 1.1rem;
			margin-bottom: 0.5rem;
		}

		.nav-card p {
			font-size: 0.8rem;
			line-height: 1.3;
		}

		.watch-card {
			display: flex;
			justify-content: center;
			align-items: end;
			width: 100vw;
			height: 100vh;
			box-sizing: border-box;
			
		}

		.watch-card .outer {
			width: 80%;
			height:35%;
		}

		.watch-card .outer .card {
			height: 100%;
		}

		.watch-card .outer .card .text {
			font-size: 2.5rem;
		}

		/* Image sizing fixes */
		.book-image {
			width: 100vw;
			height: 100vh;
			object-fit: cover;
		}

		.globe {
			width: 280px;
			height: 280px;
		}

		/* Info cards and timeline mobile fixes */
		.info-card {
			padding: 1rem;
			margin: 1rem 0;
			box-sizing: border-box;
		}

		.timeline-info {
			flex-direction: column;
			gap: 1rem;
			align-items: center;
		}

		.timeline-item {
			padding: 0.75rem;
			min-width: calc(100vw - 4rem);
			max-width: 250px;
			box-sizing: border-box;
		}

		.stats-grid {
			grid-template-columns: 1fr;
			gap: 1rem;
			margin: 1rem auto 0;
			max-width: calc(100vw - 2rem);
		}

		.stat-item {
			padding: 1rem 0.75rem;
			box-sizing: border-box;
		}

		/* Navigation and indicators mobile fixes */
		.scroll-navigation, .scroll-indicator, .scroll-text {
			visibility: hidden;
		}

		/* Mobile scroll hint 
		.scroll-container::after {
			content: "Swipe up to explore";
			position: fixed;
			bottom: 20px;
			left: 50%;
			transform: translateX(-50%);
			background: rgba(0, 0, 0, 0.7);
			color: white;
			padding: 8px 16px;
			border-radius: 20px;
			font-size: 0.8rem;
			z-index: 1000;
			animation: fade-in-out 5s ease-in-out infinite;
		}

		@keyframes fade-in-out {
			0%, 100% { opacity: 0.0; }
			75% { opacity: 0.15; }

		}*/

		/* Final CTA button fixes */
		.final-cta {
			display: flex;
			flex-direction: column;
			gap: 0.75rem;
			align-items: center;
			padding: 0 1rem;
			box-sizing: border-box;
		}

		.explore-button,
		.join-button {
			width: 100%;
			max-width: 250px;
			text-align: center;
			box-sizing: border-box;
		}

		.join-button {
			margin-left: 0;
		}

		/* Ensure all sections fit properly */
		.zoom-section {
			width: 100vw;
			box-sizing: border-box;
			padding: 0;
			margin: 0;
		}

		.zoom-content {
			width: 100%;
			box-sizing: border-box;
		}
	}

	
</style>
