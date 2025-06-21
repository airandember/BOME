<script>
  import { onMount } from 'svelte';
  import Navigation from '$lib/components/Navigation.svelte';
  import Footer from '$lib/components/Footer.svelte';
  
  let mounted = false;
  
  onMount(() => {
    mounted = true;
  });
</script>

<svelte:head>
  <title>About - BOME</title>
  <meta name="description" content="Learn about BOME's mission to provide evidence-based resources and strengthen faith through scholarly research and community engagement." />
</svelte:head>

<Navigation />

<div class="about-page">
  <div class="hero-section">
    <div class="hero-content">
      <h1>About BOME</h1>
      <p class="hero-subtitle">Building bridges between faith and knowledge through evidence-based research</p>
    </div>
  </div>

  <div class="main-content">
    <section class="mission-section">
      <div class="container">
        <div class="mission-grid">
          <div class="mission-text">
            <h2>Our Mission</h2>
            <p>
              BOME is dedicated to providing well-documented, evidence-based resources that strengthen faith 
              and enable meaningful dialogue about the Book of Mormon. Our platform serves as a bridge between 
              scholarly research and community engagement, offering tools and content that help individuals 
              explore, understand, and share knowledge with confidence and respect.
            </p>
            <p>
              We believe that faith and knowledge work hand in hand. Our ultimate goal is to bring others 
              unto Christ by providing accessible, credible resources that demonstrate the historical and 
              spiritual validity of the Book of Mormon while fostering an environment of learning and growth.
            </p>
          </div>
          <div class="mission-visual">
            <div class="mission-card">
              <div class="icon">🔍</div>
              <h3>Evidence-Based Research</h3>
              <p>Rigorous scholarship using multiple disciplines</p>
            </div>
            <div class="mission-card">
              <div class="icon">🤝</div>
              <h3>Community Engagement</h3>
              <p>Fostering meaningful dialogue and learning</p>
            </div>
            <div class="mission-card">
              <div class="icon">✨</div>
              <h3>Faith Strengthening</h3>
              <p>Supporting spiritual growth through knowledge</p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <section class="approach-section">
      <div class="container">
        <h2>Our Approach</h2>
        <div class="approach-grid">
          <div class="approach-item">
            <div class="approach-number">01</div>
            <h3>Scholarly Excellence</h3>
            <p>
              We maintain the highest standards of academic rigor, utilizing peer-reviewed research 
              and credible sources across multiple disciplines including archaeology, linguistics, 
              genetics, and historical analysis.
            </p>
          </div>
          <div class="approach-item">
            <div class="approach-number">02</div>
            <h3>Balanced Perspective</h3>
            <p>
              We strive to present information objectively, acknowledging both supporting evidence 
              and areas of ongoing research, allowing individuals to form their own informed conclusions.
            </p>
          </div>
          <div class="approach-item">
            <div class="approach-number">03</div>
            <h3>Accessible Learning</h3>
            <p>
              Complex research is made accessible through clear explanations, visual aids, and 
              multimedia content that serves both scholarly and general audiences.
            </p>
          </div>
          <div class="approach-item">
            <div class="approach-number">04</div>
            <h3>Community Driven</h3>
            <p>
              Our platform thrives on community participation, encouraging discussion, questions, 
              and collaborative learning among members with diverse backgrounds and perspectives.
            </p>
          </div>
        </div>
      </div>
    </section>

    <section class="research-areas">
      <div class="container">
        <h2>Areas of Research</h2>
        <div class="research-grid">
          <div class="research-category">
            <h3>Archaeological Evidence</h3>
            <ul>
              <li>Ancient civilizations and settlements</li>
              <li>Metallurgy and technology</li>
              <li>Cultural practices and artifacts</li>
              <li>Geographic correlations</li>
            </ul>
          </div>
          <div class="research-category">
            <h3>Linguistic Analysis</h3>
            <ul>
              <li>Hebrew and Semitic language patterns</li>
              <li>Ancient writing systems</li>
              <li>Translation methodologies</li>
              <li>Comparative textual analysis</li>
            </ul>
          </div>
          <div class="research-category">
            <h3>Historical Context</h3>
            <ul>
              <li>Ancient Near Eastern studies</li>
              <li>Biblical scholarship connections</li>
              <li>Cultural and religious practices</li>
              <li>Historical timeline correlations</li>
            </ul>
          </div>
          <div class="research-category">
            <h3>Scientific Studies</h3>
            <ul>
              <li>DNA and genetic research</li>
              <li>Climate and environmental data</li>
              <li>Astronomical observations</li>
              <li>Anthropological findings</li>
            </ul>
          </div>
        </div>
      </div>
    </section>

    <section class="values-section">
      <div class="container">
        <h2>Our Core Values</h2>
        <div class="values-grid">
          <div class="value-item">
            <div class="value-icon">🎯</div>
            <h3>Truth Seeking</h3>
            <p>We are committed to honest inquiry and the pursuit of truth, wherever it may lead.</p>
          </div>
          <div class="value-item">
            <div class="value-icon">🤲</div>
            <h3>Respectful Dialogue</h3>
            <p>We foster an environment of mutual respect and constructive conversation.</p>
          </div>
          <div class="value-item">
            <div class="value-icon">📚</div>
            <h3>Continuous Learning</h3>
            <p>We embrace ongoing education and remain open to new discoveries and insights.</p>
          </div>
          <div class="value-item">
            <div class="value-icon">🌟</div>
            <h3>Faith & Reason</h3>
            <p>We believe that faith and reason complement each other in the search for truth.</p>
          </div>
        </div>
      </div>
    </section>

    <section class="community-section">
      <div class="container">
        <h2>Join Our Community</h2>
        <div class="community-content">
          <div class="community-text">
            <p>
              BOME is more than a platform—it's a community of seekers, scholars, and believers 
              united in the pursuit of knowledge and truth. Whether you're a researcher, student, 
              educator, or simply someone with questions, you'll find a welcoming environment 
              where your contributions are valued.
            </p>
            <p>
              Our community includes members from diverse backgrounds, professions, and levels 
              of expertise, all working together to strengthen understanding and faith through 
              careful study and meaningful dialogue.
            </p>
          </div>
          <div class="community-stats">
            <div class="stat-item">
              <div class="stat-number">10K+</div>
              <div class="stat-label">Active Members</div>
            </div>
            <div class="stat-item">
              <div class="stat-number">500+</div>
              <div class="stat-label">Research Articles</div>
            </div>
            <div class="stat-item">
              <div class="stat-number">50+</div>
              <div class="stat-label">Expert Contributors</div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <section class="get-involved">
      <div class="container">
        <h2>Get Involved</h2>
        <div class="involvement-grid">
          <div class="involvement-card">
            <h3>Contribute Research</h3>
            <p>Share your scholarly work and insights with our community of researchers and learners.</p>
            <a href="/dashboard" class="cta-button">Start Contributing</a>
          </div>
          <div class="involvement-card">
            <h3>Join Discussions</h3>
            <p>Participate in meaningful conversations about evidence, faith, and discovery.</p>
            <a href="/register" class="cta-button">Join Community</a>
          </div>
          <div class="involvement-card">
            <h3>Support Our Mission</h3>
            <p>Help us continue providing valuable resources through your support and engagement.</p>
            <a href="/subscription" class="cta-button">Learn More</a>
          </div>
        </div>
      </div>
    </section>

    <section class="quote-section">
      <div class="container">
        <blockquote>
          <p>
            "Truth will prevail. It always has and it always will. Our responsibility is to seek it 
            diligently, share it respectfully, and live it faithfully."
          </p>
          <cite>— BOME Community Principle</cite>
        </blockquote>
      </div>
    </section>
  </div>
</div>

<Footer />

<style>
  .about-page {
    min-height: 100vh;
    background: var(--bg-primary);
    color: var(--text-primary);
    padding-top: 80px; /* Account for fixed navigation */
  }

  .hero-section {
    padding: var(--space-4xl) 0;
    text-align: center;
    background: var(--primary-gradient);
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
    background: var(--bg-glass);
    backdrop-filter: blur(10px);
  }

  .hero-content {
    position: relative;
    z-index: 2;
  }

  .hero-content h1 {
    font-size: var(--text-5xl);
    font-weight: 700;
    color: var(--text-inverse);
    margin-bottom: var(--space-lg);
    text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  .hero-subtitle {
    font-size: var(--text-xl);
    color: rgba(255, 255, 255, 0.9);
    max-width: 600px;
    margin: 0 auto;
    line-height: var(--leading-relaxed);
  }

  .main-content {
    background: var(--bg-primary);
  }

  section {
    padding: var(--space-4xl) 0;
  }

  .mission-section {
    background: var(--bg-secondary);
  }

  .mission-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-3xl);
    align-items: center;
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 var(--space-xl);
  }

  .mission-text h2 {
    font-size: var(--text-4xl);
    margin-bottom: var(--space-xl);
    color: var(--primary);
  }

  .mission-text p {
    font-size: var(--text-lg);
    line-height: var(--leading-relaxed);
    margin-bottom: var(--space-lg);
    color: var(--text-secondary);
  }

  .mission-visual {
    display: grid;
    gap: var(--space-lg);
  }

  .mission-card {
    background: var(--bg-primary);
    border: 1px solid var(--gray-200);
    border-radius: var(--radius-xl);
    padding: var(--space-xl);
    text-align: center;
    transition: all var(--transition-normal);
    box-shadow: var(--shadow-md);
  }

  .mission-card:hover {
    transform: translateY(-5px);
    box-shadow: var(--shadow-xl);
    border-color: var(--primary);
  }

  .mission-card .icon {
    font-size: var(--text-4xl);
    margin-bottom: var(--space-md);
  }

  .mission-card h3 {
    font-size: var(--text-xl);
    margin-bottom: var(--space-sm);
    color: var(--primary);
  }

  .approach-section {
    background: var(--bg-primary);
  }

  .approach-section h2 {
    text-align: center;
    font-size: var(--text-4xl);
    margin-bottom: var(--space-2xl);
    color: var(--primary);
    max-width: 1200px;
    margin-left: auto;
    margin-right: auto;
    padding: 0 var(--space-xl);
  }

  .approach-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: var(--space-xl);
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 var(--space-xl);
  }

  .approach-item {
    background: var(--bg-primary);
    border: 1px solid var(--gray-200);
    border-radius: var(--radius-xl);
    padding: var(--space-xl);
    position: relative;
    transition: all var(--transition-normal);
    box-shadow: var(--shadow-md);
  }

  .approach-item:hover {
    transform: translateY(-5px);
    box-shadow: var(--shadow-xl);
    border-color: var(--primary);
  }

  .approach-number {
    position: absolute;
    top: -15px;
    left: var(--space-xl);
    background: var(--primary);
    color: var(--text-inverse);
    width: 40px;
    height: 40px;
    border-radius: var(--radius-full);
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 700;
    font-size: var(--text-lg);
    box-shadow: var(--shadow-md);
  }

  .approach-item h3 {
    margin: var(--space-md) 0;
    color: var(--primary);
    font-size: var(--text-xl);
  }

  .approach-item p {
    line-height: var(--leading-relaxed);
    color: var(--text-secondary);
  }

  .research-areas {
    background: var(--bg-secondary);
  }

  .research-areas h2 {
    text-align: center;
    font-size: var(--text-4xl);
    margin-bottom: var(--space-2xl);
    color: var(--primary);
    max-width: 1200px;
    margin-left: auto;
    margin-right: auto;
    padding: 0 var(--space-xl);
  }

  .research-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: var(--space-xl);
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 var(--space-xl);
  }

  .research-category {
    background: var(--bg-primary);
    border: 1px solid var(--gray-200);
    border-radius: var(--radius-xl);
    padding: var(--space-xl);
    transition: all var(--transition-normal);
    box-shadow: var(--shadow-md);
  }

  .research-category:hover {
    transform: translateY(-5px);
    box-shadow: var(--shadow-xl);
    border-color: var(--primary);
  }

  .research-category h3 {
    margin-bottom: var(--space-lg);
    color: var(--primary);
    font-size: var(--text-xl);
  }

  .research-category ul {
    list-style: none;
    padding: 0;
  }

  .research-category li {
    padding: var(--space-sm) 0;
    color: var(--text-secondary);
    position: relative;
    padding-left: var(--space-lg);
  }

  .research-category li::before {
    content: "▸";
    position: absolute;
    left: 0;
    color: var(--primary);
    font-weight: 700;
  }

  .values-section {
    background: var(--bg-primary);
  }

  .values-section h2 {
    text-align: center;
    font-size: var(--text-4xl);
    margin-bottom: var(--space-2xl);
    color: var(--primary);
    max-width: 1200px;
    margin-left: auto;
    margin-right: auto;
    padding: 0 var(--space-xl);
  }

  .values-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: var(--space-xl);
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 var(--space-xl);
  }

  .value-item {
    text-align: center;
    padding: var(--space-xl);
    background: var(--bg-primary);
    border: 1px solid var(--gray-200);
    border-radius: var(--radius-xl);
    transition: all var(--transition-normal);
    box-shadow: var(--shadow-md);
  }

  .value-item:hover {
    transform: translateY(-5px);
    box-shadow: var(--shadow-xl);
    border-color: var(--primary);
  }

  .value-icon {
    font-size: var(--text-5xl);
    margin-bottom: var(--space-md);
  }

  .value-item h3 {
    margin-bottom: var(--space-md);
    color: var(--primary);
    font-size: var(--text-xl);
  }

  .value-item p {
    line-height: var(--leading-relaxed);
    color: var(--text-secondary);
  }

  .community-section {
    background: var(--bg-secondary);
  }

  .community-section h2 {
    text-align: center;
    font-size: var(--text-4xl);
    margin-bottom: var(--space-2xl);
    color: var(--primary);
    max-width: 1200px;
    margin-left: auto;
    margin-right: auto;
    padding: 0 var(--space-xl);
  }

  .community-content {
    display: grid;
    grid-template-columns: 2fr 1fr;
    gap: var(--space-3xl);
    align-items: center;
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 var(--space-xl);
  }

  .community-text p {
    font-size: var(--text-lg);
    line-height: var(--leading-relaxed);
    margin-bottom: var(--space-lg);
    color: var(--text-secondary);
  }

  .community-stats {
    display: grid;
    gap: var(--space-xl);
  }

  .stat-item {
    text-align: center;
    padding: var(--space-lg);
    background: var(--bg-primary);
    border: 1px solid var(--gray-200);
    border-radius: var(--radius-xl);
    box-shadow: var(--shadow-md);
    transition: all var(--transition-normal);
  }

  .stat-item:hover {
    transform: translateY(-2px);
    box-shadow: var(--shadow-lg);
    border-color: var(--primary);
  }

  .stat-number {
    font-size: var(--text-4xl);
    font-weight: 700;
    color: var(--primary);
    margin-bottom: var(--space-sm);
  }

  .stat-label {
    color: var(--text-secondary);
    font-size: var(--text-sm);
    text-transform: uppercase;
    letter-spacing: 1px;
    font-weight: 600;
  }

  .get-involved {
    background: var(--bg-primary);
  }

  .get-involved h2 {
    text-align: center;
    font-size: var(--text-4xl);
    margin-bottom: var(--space-2xl);
    color: var(--primary);
    max-width: 1200px;
    margin-left: auto;
    margin-right: auto;
    padding: 0 var(--space-xl);
  }

  .involvement-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: var(--space-xl);
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 var(--space-xl);
  }

  .involvement-card {
    background: var(--bg-primary);
    border: 1px solid var(--gray-200);
    border-radius: var(--radius-xl);
    padding: var(--space-xl);
    text-align: center;
    transition: all var(--transition-normal);
    box-shadow: var(--shadow-md);
  }

  .involvement-card:hover {
    transform: translateY(-5px);
    box-shadow: var(--shadow-xl);
    border-color: var(--primary);
  }

  .involvement-card h3 {
    margin-bottom: var(--space-md);
    color: var(--primary);
    font-size: var(--text-xl);
  }

  .involvement-card p {
    margin-bottom: var(--space-xl);
    line-height: var(--leading-relaxed);
    color: var(--text-secondary);
  }

  .cta-button {
    display: inline-block;
    background: var(--primary);
    color: var(--text-inverse);
    padding: var(--space-md) var(--space-xl);
    border-radius: var(--radius-md);
    text-decoration: none;
    font-weight: 600;
    transition: all var(--transition-normal);
    box-shadow: var(--shadow-md);
  }

  .cta-button:hover {
    background: var(--primary-dark);
    transform: translateY(-2px);
    box-shadow: var(--shadow-lg);
  }

  .quote-section {
    background: var(--bg-secondary);
    text-align: center;
  }

  blockquote {
    max-width: 800px;
    margin: 0 auto;
    padding: var(--space-2xl);
    background: var(--bg-primary);
    border: 1px solid var(--gray-200);
    border-radius: var(--radius-xl);
    position: relative;
    box-shadow: var(--shadow-lg);
  }

  blockquote::before {
    content: '"';
    font-size: var(--text-6xl);
    color: var(--primary);
    position: absolute;
    top: -1rem;
    left: var(--space-xl);
    line-height: 1;
  }

  blockquote p {
    font-size: var(--text-xl);
    line-height: var(--leading-relaxed);
    color: var(--text-secondary);
    margin-bottom: var(--space-md);
    font-style: italic;
  }

  blockquote cite {
    color: var(--primary);
    font-weight: 600;
    font-style: normal;
  }

  @media (max-width: 768px) {
    .about-page {
      padding-top: 60px;
    }

    .hero-content h1 {
      font-size: var(--text-4xl);
    }

    .hero-subtitle {
      font-size: var(--text-lg);
    }

    .mission-grid,
    .community-content {
      grid-template-columns: 1fr;
      gap: var(--space-xl);
      padding: 0 var(--space-md);
    }

    .approach-grid,
    .research-grid,
    .values-grid,
    .involvement-grid {
      grid-template-columns: 1fr;
      padding: 0 var(--space-md);
    }

    .approach-section h2,
    .research-areas h2,
    .values-section h2,
    .get-involved h2,
    .community-section h2 {
      padding: 0 var(--space-md);
    }

    section {
      padding: var(--space-2xl) 0;
    }

    blockquote {
      padding: var(--space-xl);
      margin: 0 var(--space-md);
    }

    blockquote p {
      font-size: var(--text-lg);
    }

    blockquote::before {
      font-size: var(--text-5xl);
      left: var(--space-md);
    }
  }
</style> 