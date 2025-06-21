<script>
  import { onMount } from 'svelte';
  import Navigation from '$lib/components/Navigation.svelte';
  import Footer from '$lib/components/Footer.svelte';
  
  let mounted = false;
  let formData = {
    name: '',
    email: '',
    subject: '',
    category: 'general',
    message: ''
  };
  let isSubmitting = false;
  let submitStatus = '';

  onMount(() => {
    mounted = true;
  });

  const handleSubmit = async (e) => {
    e.preventDefault();
    isSubmitting = true;
    submitStatus = '';

    try {
      // Simulate form submission - replace with actual API call
      await new Promise(resolve => setTimeout(resolve, 1500));
      
      // Reset form
      formData = {
        name: '',
        email: '',
        subject: '',
        category: 'general',
        message: ''
      };
      
      submitStatus = 'success';
    } catch (error) {
      submitStatus = 'error';
    } finally {
      isSubmitting = false;
    }
  };
</script>

<svelte:head>
  <title>Contact Us - BOME</title>
  <meta name="description" content="Get in touch with the BOME team. We're here to help with questions about Book of Mormon evidence, research, technical support, and more." />
</svelte:head>

<Navigation />

<div class="contact-page">
  <div class="hero-section">
    <div class="hero-content">
      <h1>Contact Us</h1>
      <p class="hero-subtitle">We're here to help and answer any questions you might have</p>
    </div>
  </div>

  <div class="main-content">
    <!-- Contact Information Section -->
    <section class="contact-info-section">
      <div class="container">
        <div class="contact-grid">
          <div class="contact-methods">
            <h2>Get in Touch</h2>
            <p>
              Whether you have questions about our research, need technical support, or want to 
              contribute to our mission, we'd love to hear from you. Choose the best way to reach us.
            </p>

            <div class="contact-cards">
              <div class="contact-card">
                <div class="contact-icon">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path>
                    <polyline points="22,6 12,13 2,6"></polyline>
                  </svg>
                </div>
                <h3>Email Us</h3>
                <p>General inquiries and support</p>
                <a href="mailto:contact@bome.org" class="contact-link">contact@bome.org</a>
              </div>

              <div class="contact-card">
                <div class="contact-icon">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M9 11a3 3 0 1 0 6 0a3 3 0 0 0 -6 0"></path>
                    <path d="M17.657 16.657l-4.243 4.243a2 2 0 0 1 -2.827 0l-4.244 -4.243a8 8 0 1 1 11.314 0z"></path>
                  </svg>
                </div>
                <h3>Office Location</h3>
                <p>Visit us during business hours</p>
                <address class="contact-address">
                  123 Research Way<br>
                  Academic District<br>
                  Salt Lake City, UT 84101
                </address>
              </div>

              <div class="contact-card">
                <div class="contact-icon">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"></path>
                  </svg>
                </div>
                <h3>Phone Support</h3>
                <p>Monday - Friday, 9 AM - 5 PM MST</p>
                <a href="tel:+1-801-555-0123" class="contact-link">+1 (801) 555-0123</a>
              </div>

              <div class="contact-card">
                <div class="contact-icon">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"></path>
                  </svg>
                </div>
                <h3>Live Chat</h3>
                <p>Real-time support available</p>
                <button class="contact-link chat-button">Start Chat</button>
              </div>
            </div>
          </div>

          <!-- Contact Form -->
          <div class="contact-form-container">
            <form class="contact-form" on:submit={handleSubmit}>
              <h3>Send us a Message</h3>
              
              <div class="form-group">
                <label for="name">Full Name *</label>
                <input 
                  type="text" 
                  id="name" 
                  bind:value={formData.name}
                  required 
                  placeholder="Enter your full name"
                />
              </div>

              <div class="form-group">
                <label for="email">Email Address *</label>
                <input 
                  type="email" 
                  id="email" 
                  bind:value={formData.email}
                  required 
                  placeholder="Enter your email address"
                />
              </div>

              <div class="form-group">
                <label for="category">Category</label>
                <select id="category" bind:value={formData.category}>
                  <option value="general">General Inquiry</option>
                  <option value="research">Research Question</option>
                  <option value="technical">Technical Support</option>
                  <option value="partnership">Partnership</option>
                  <option value="media">Media Inquiry</option>
                  <option value="feedback">Feedback</option>
                </select>
              </div>

              <div class="form-group">
                <label for="subject">Subject *</label>
                <input 
                  type="text" 
                  id="subject" 
                  bind:value={formData.subject}
                  required 
                  placeholder="Brief description of your inquiry"
                />
              </div>

              <div class="form-group">
                <label for="message">Message *</label>
                <textarea 
                  id="message" 
                  bind:value={formData.message}
                  required 
                  placeholder="Please provide details about your inquiry..."
                  rows="6"
                ></textarea>
              </div>

              <button 
                type="submit" 
                class="submit-button"
                disabled={isSubmitting}
              >
                {#if isSubmitting}
                  <div class="spinner"></div>
                  Sending...
                {:else}
                  Send Message
                {/if}
              </button>

              {#if submitStatus === 'success'}
                <div class="status-message success">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="20,6 9,17 4,12"></polyline>
                  </svg>
                  Thank you! Your message has been sent successfully. We'll get back to you soon.
                </div>
              {/if}

              {#if submitStatus === 'error'}
                <div class="status-message error">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"></circle>
                    <line x1="15" y1="9" x2="9" y2="15"></line>
                    <line x1="9" y1="9" x2="15" y2="15"></line>
                  </svg>
                  Sorry, there was an error sending your message. Please try again or contact us directly.
                </div>
              {/if}
            </form>
          </div>
        </div>
      </div>
    </section>

    <!-- FAQ Section -->
    <section class="faq-section">
      <div class="container">
        <h2>Frequently Asked Questions</h2>
        <div class="faq-grid">
          <div class="faq-item">
            <h3>How can I contribute research to BOME?</h3>
            <p>
              We welcome scholarly contributions from researchers and academics. Please contact us 
              with your research proposal, and our editorial team will review it for inclusion.
            </p>
          </div>

          <div class="faq-item">
            <h3>Is BOME content peer-reviewed?</h3>
            <p>
              Yes, all research content on BOME undergoes a rigorous peer-review process by our 
              panel of expert scholars across multiple disciplines.
            </p>
          </div>

          <div class="faq-item">
            <h3>How do I report technical issues?</h3>
            <p>
              For technical support, please use the "Technical Support" category in our contact 
              form or email us directly at support@bome.org with details about the issue.
            </p>
          </div>

          <div class="faq-item">
            <h3>Can I use BOME content for educational purposes?</h3>
            <p>
              Yes, BOME content is available for educational use. Please review our terms of use 
              and provide proper attribution when using our materials.
            </p>
          </div>

          <div class="faq-item">
            <h3>How do I become a premium member?</h3>
            <p>
              Visit our subscription page to learn about premium membership benefits and sign up 
              for access to exclusive content and features.
            </p>
          </div>

          <div class="faq-item">
            <h3>Do you offer group or institutional subscriptions?</h3>
            <p>
              Yes, we offer special pricing for educational institutions and organizations. 
              Contact us for more information about group subscriptions.
            </p>
          </div>
        </div>
      </div>
    </section>

    <!-- Office Hours Section -->
    <section class="hours-section">
      <div class="container">
        <div class="hours-content">
          <div class="hours-info">
            <h2>Office Hours & Response Times</h2>
            <div class="hours-grid">
              <div class="hours-item">
                <h3>Phone Support</h3>
                <p>Monday - Friday: 9:00 AM - 5:00 PM MST</p>
                <p>Saturday: 10:00 AM - 2:00 PM MST</p>
                <p>Sunday: Closed</p>
              </div>

              <div class="hours-item">
                <h3>Email Response</h3>
                <p>General Inquiries: Within 24 hours</p>
                <p>Technical Support: Within 4 hours</p>
                <p>Research Questions: Within 48 hours</p>
              </div>

              <div class="hours-item">
                <h3>Live Chat</h3>
                <p>Available during phone support hours</p>
                <p>Average response time: Under 5 minutes</p>
                <p>After hours: Leave a message</p>
              </div>
            </div>
          </div>

          <div class="emergency-contact">
            <h3>Emergency Contact</h3>
            <p>
              For urgent technical issues affecting the platform, please call our emergency line:
            </p>
            <a href="tel:+1-801-555-0911" class="emergency-number">+1 (801) 555-0911</a>
            <p class="emergency-note">Available 24/7 for critical issues only</p>
          </div>
        </div>
      </div>
    </section>
  </div>
</div>

<Footer />

<style>
  .contact-page {
    min-height: 100vh;
    background: var(--bg-primary);
    color: var(--text-primary);
    padding-top: 80px;
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

  .container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 var(--space-xl);
  }

  section {
    padding: var(--space-4xl) 0;
  }

  .contact-info-section {
    background: var(--bg-secondary);
  }

  .contact-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-3xl);
    align-items: start;
  }

  .contact-methods h2 {
    font-size: var(--text-4xl);
    margin-bottom: var(--space-lg);
    color: var(--primary);
  }

  .contact-methods p {
    font-size: var(--text-lg);
    line-height: var(--leading-relaxed);
    margin-bottom: var(--space-xl);
    color: var(--text-secondary);
  }

  .contact-cards {
    display: grid;
    gap: var(--space-lg);
  }

  .contact-card {
    background: var(--bg-primary);
    border: 1px solid var(--gray-200);
    border-radius: var(--radius-xl);
    padding: var(--space-xl);
    transition: all var(--transition-normal);
    box-shadow: var(--shadow-md);
  }

  .contact-card:hover {
    transform: translateY(-2px);
    box-shadow: var(--shadow-lg);
    border-color: var(--primary);
  }

  .contact-icon {
    width: 48px;
    height: 48px;
    background: var(--primary);
    border-radius: var(--radius-lg);
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: var(--space-md);
  }

  .contact-icon svg {
    width: 24px;
    height: 24px;
    color: var(--text-inverse);
  }

  .contact-card h3 {
    font-size: var(--text-xl);
    margin-bottom: var(--space-sm);
    color: var(--primary);
  }

  .contact-card p {
    color: var(--text-secondary);
    margin-bottom: var(--space-md);
  }

  .contact-link {
    color: var(--primary);
    text-decoration: none;
    font-weight: 600;
    transition: color var(--transition-fast);
  }

  .contact-link:hover {
    color: var(--primary-dark);
  }

  .chat-button {
    background: none;
    border: none;
    padding: 0;
    font: inherit;
    cursor: pointer;
  }

  .contact-address {
    font-style: normal;
    line-height: var(--leading-relaxed);
    color: var(--text-secondary);
  }

  .contact-form-container {
    background: var(--bg-primary);
    border: 1px solid var(--gray-200);
    border-radius: var(--radius-xl);
    padding: var(--space-2xl);
    box-shadow: var(--shadow-lg);
  }

  .contact-form h3 {
    font-size: var(--text-2xl);
    margin-bottom: var(--space-xl);
    color: var(--primary);
  }

  .form-group {
    margin-bottom: var(--space-lg);
  }

  .form-group label {
    display: block;
    margin-bottom: var(--space-sm);
    font-weight: 600;
    color: var(--text-primary);
  }

  .form-group input,
  .form-group select,
  .form-group textarea {
    width: 100%;
    padding: var(--space-md);
    border: 1px solid var(--gray-300);
    border-radius: var(--radius-md);
    font-size: var(--text-base);
    transition: border-color var(--transition-fast);
    background: var(--bg-primary);
    color: var(--text-primary);
  }

  .form-group input:focus,
  .form-group select:focus,
  .form-group textarea:focus {
    outline: none;
    border-color: var(--primary);
    box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
  }

  .form-group textarea {
    resize: vertical;
    min-height: 120px;
  }

  .submit-button {
    background: var(--primary);
    color: var(--text-inverse);
    border: none;
    padding: var(--space-md) var(--space-xl);
    border-radius: var(--radius-md);
    font-size: var(--text-base);
    font-weight: 600;
    cursor: pointer;
    transition: all var(--transition-normal);
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    box-shadow: var(--shadow-md);
  }

  .submit-button:hover:not(:disabled) {
    background: var(--primary-dark);
    transform: translateY(-2px);
    box-shadow: var(--shadow-lg);
  }

  .submit-button:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }

  .spinner {
    width: 16px;
    height: 16px;
    border: 2px solid transparent;
    border-top: 2px solid currentColor;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .status-message {
    margin-top: var(--space-lg);
    padding: var(--space-md);
    border-radius: var(--radius-md);
    display: flex;
    align-items: center;
    gap: var(--space-sm);
  }

  .status-message.success {
    background: rgba(0, 212, 170, 0.1);
    color: var(--success);
    border: 1px solid rgba(0, 212, 170, 0.2);
  }

  .status-message.error {
    background: rgba(255, 107, 107, 0.1);
    color: var(--error);
    border: 1px solid rgba(255, 107, 107, 0.2);
  }

  .status-message svg {
    width: 20px;
    height: 20px;
    flex-shrink: 0;
  }

  .faq-section {
    background: var(--bg-primary);
  }

  .faq-section h2 {
    text-align: center;
    font-size: var(--text-4xl);
    margin-bottom: var(--space-2xl);
    color: var(--primary);
  }

  .faq-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
    gap: var(--space-xl);
  }

  .faq-item {
    background: var(--bg-primary);
    border: 1px solid var(--gray-200);
    border-radius: var(--radius-xl);
    padding: var(--space-xl);
    transition: all var(--transition-normal);
    box-shadow: var(--shadow-md);
  }

  .faq-item:hover {
    transform: translateY(-2px);
    box-shadow: var(--shadow-lg);
    border-color: var(--primary);
  }

  .faq-item h3 {
    font-size: var(--text-lg);
    margin-bottom: var(--space-md);
    color: var(--primary);
  }

  .faq-item p {
    color: var(--text-secondary);
    line-height: var(--leading-relaxed);
  }

  .hours-section {
    background: var(--bg-secondary);
  }

  .hours-content {
    display: grid;
    grid-template-columns: 2fr 1fr;
    gap: var(--space-3xl);
    align-items: start;
  }

  .hours-info h2 {
    font-size: var(--text-4xl);
    margin-bottom: var(--space-xl);
    color: var(--primary);
  }

  .hours-grid {
    display: grid;
    gap: var(--space-xl);
  }

  .hours-item {
    background: var(--bg-primary);
    border: 1px solid var(--gray-200);
    border-radius: var(--radius-xl);
    padding: var(--space-xl);
    box-shadow: var(--shadow-md);
  }

  .hours-item h3 {
    font-size: var(--text-xl);
    margin-bottom: var(--space-md);
    color: var(--primary);
  }

  .hours-item p {
    color: var(--text-secondary);
    margin-bottom: var(--space-sm);
  }

  .emergency-contact {
    background: var(--bg-primary);
    border: 1px solid var(--error);
    border-radius: var(--radius-xl);
    padding: var(--space-xl);
    text-align: center;
    box-shadow: var(--shadow-md);
  }

  .emergency-contact h3 {
    font-size: var(--text-xl);
    margin-bottom: var(--space-md);
    color: var(--error);
  }

  .emergency-contact p {
    color: var(--text-secondary);
    margin-bottom: var(--space-md);
  }

  .emergency-number {
    display: block;
    font-size: var(--text-2xl);
    font-weight: 700;
    color: var(--error);
    text-decoration: none;
    margin: var(--space-md) 0;
  }

  .emergency-note {
    font-size: var(--text-sm);
    color: var(--text-muted);
    font-style: italic;
  }

  @media (max-width: 768px) {
    .contact-page {
      padding-top: 60px;
    }

    .hero-content h1 {
      font-size: var(--text-4xl);
    }

    .hero-subtitle {
      font-size: var(--text-lg);
    }

    .container {
      padding: 0 var(--space-md);
    }

    .contact-grid,
    .hours-content {
      grid-template-columns: 1fr;
      gap: var(--space-xl);
    }

    .faq-grid {
      grid-template-columns: 1fr;
    }

    section {
      padding: var(--space-2xl) 0;
    }

    .contact-form-container {
      padding: var(--space-xl);
    }
  }
</style> 