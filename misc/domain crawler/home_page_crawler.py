from urllib.parse import urlparse, urljoin
import requests
from bs4 import BeautifulSoup
import openai
import time


# Other files and helper functions
from find_pricing import find_pricing_elements
from find_CTA import find_main_CTAs
from find_customer_base import find_main_customer_base
from find_ideal_buying_companies import find_main_buying_companies
from quick_glance import find_quick_glance
from fit_assessment import find_possible_services
from find_category import find_exact_category
from find_logos import find_logos_and_testimonials
from content_piece_analytics import find_content_metrics

# Set up OpenAI API credentials
openai.api_key = ''

# List of domains
domains = [
    'zapscale.com',
    
    # Add more domains here
]

def format_text(soup_text):
    formatted_text = ' '.join(soup_text.split())
    return formatted_text

def get_main_links(soup, homepage_url):    
    main_links = []

    # Find all anchor tags on the homepage
    anchor_tags = soup.find_all('a')
    
    # Extract the absolute URLs of the main links
    for tag in anchor_tags:
        link = tag.get('href')
        if link:
            absolute_url = urljoin(homepage_url, link)
            main_links.append(absolute_url)
    
    # Filter out any duplicate or empty links
    main_links = list(set(main_links))
    main_links = [link for link in main_links if link and urlparse(link).scheme]
    
    return main_links

# Function to scrape information from a URL
def scrape_url(url):
    response = requests.get(url)
    soup = BeautifulSoup(response.content, 'html.parser')    

    # Extract relevant information from the webpage
    # Modify this function based on the structure of the webpages you're scraping
    # Return a dictionary containing the extracted information

    info = {
        'title': soup.title.text,
        'main_l1_pages': get_main_links(soup, homepage_url=url),
        'text': format_text(soup.get_text()),  # Example: Extracting all text from the webpage
        'main_headline': soup.find("h1").text
    }

    return info



# Iterate over each domain
for domain in domains:
    # Scrape information from the domain's homepage
    domain_info = scrape_url(f"http://{domain}")

    ## AT THIS POINT, WE HAVE THE HTML TEXT FOR THE DOMAIN

    # # print('Domain INFO')
    # print(domain_info['main_l1_pages'])
    print('')

    # # # QUICK GLANCE
    quick_glance = find_quick_glance(domain_info['text'])

    time.sleep(5)

    # # # # FIND IDEAL COMPANY
    main_buyer_companies = find_main_buying_companies(domain_info['text'])

    time.sleep(5)

    # # # # FIND MAIN CUSTOMER BASE
    main_customer_base = find_main_customer_base(domain_info['text'])

    time.sleep(5)

    # # # # FIND CTAs
    main_ctas = find_main_CTAs(domain_info['text'])

    time.sleep(5)

    # # # # FIND PRICING STRUCTURE AND BUCKETS
    # # # # temp_li_list = ['http://factors.ai/solutions/cmo', 'http://factors.ai/solutions/marketing-ops', 'http://factors.ai', 'http://factors.ai/blog/saas-marketing-reporting', 'http://factors.ai/ama/google-ads-qna-with-ashwin-and-rahul', 'http://factors.ai/solutions/content-marketer', 'http://factors.ai/privacy-policy', 'http://factors.ai/customers', 'http://factors.ai/blog/cmo-responsibilities', 'https://twitter.com/factorsai?lang=en', 'http://factors.ai/ama/digital-trends-in-practice-with-ashit-malik', 'https://www.factors.ai/sitemap.xml', 'http://factors.ai/terms-of-use', 'http://factors.ai/conversations/siddharth-deswal-making-marketing-operations-analytics-simple-for-a-b2b-saas-startup', 'http://factors.ai/press', 'http://factors.ai/podcast/all-things-intent-signals-with-monish-munshi', 'https://www.facebook.com/factorsai/', 'http://factors.ai/conversations/siddharth-sharma-marketing-analytics-decision-making-for-a-saas-startup', 'http://factors.ai/podcast/digital-marketing-levers-for-post-product-market-fit-firms-with-shiyam-sunder', 'http://factors.ai/podcast/revenue-marketing-and-more-with-alex-sofronas', 'http://factors.ai/case-studies', 'http://factors.ai/labs/do-leads-revisit-your-website-after-a-demo', 'http://factors.ai/company', 'https://www.factors.ai/blog/6sense-factors-ai-partnership-announcement', 'http://factors.ai/podcast', 'http://factors.ai/pricing', 'http://factors.ai/ama', 'http://factors.ai/features', 'http://factors.ai/labs', 'https://app.factors.ai/', 'http://factors.ai/data-processing-agreement', 'https://in.linkedin.com/company/factors-ai', 'http://factors.ai/labs/your-form-fields-matter-heres-how', 'mailto:solutions@factors.ai', 'https://help.factors.ai/en/', 'https://www.producthunt.com/golden-kitty-awards-2021/ai-machine-learning?bc=1', 'http://factors.ai/blog', 'http://factors.ai/library', 'https://6sense.com/', 'http://factors.ai/', 'http://factors.ai/solutions/demand-generation', 'http://factors.ai/blog/marketing-touchpoints-to-guide-demo-bookings', 'http://factors.ai/security', 'http://factors.ai/conversations', 'http://factors.ai/conversations/paresh-mandhyan-making-b2b-marketing-measurable-and-data-driven-in-2021', 'http://factors.ai/ama/befriend-your-data-with-aravind-murthy-anshul-jain-and-vinith-kumar', 'https://app.factors.ai/signup', 'http://factors.ai/labs/whats-the-right-number-of-demo-form-fields', 'https://angel.co/company/factors-ai/jobs']
    pricing_elements = find_pricing_elements(domain_info['main_l1_pages'])

    time.sleep(5)

    # # # # POSSIBLE SERVICES
    possible_services = find_possible_services(domain_info['text'])

    time.sleep(5)

    # # # Categorization
    categorization = find_exact_category(domain_info['text'])


    # ## Logos and testimonials
    logos = find_logos_and_testimonials(domain_info['text'])


    ## Content Analytics
    # date_published = find_content_metrics(domain_info['text'])




    
