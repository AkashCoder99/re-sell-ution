/**
 * Cypress E2E — Search Listings page
 * Covers search input, keyword search, and result display.
 */
describe('Search Listings', () => {
  beforeEach(() => {
    cy.enterPreview()
    cy.contains('button', 'Search by Keyword').click()
    cy.wait(600)
  })

  it('shows the Search heading', () => {
    cy.contains('Search').should('be.visible')
    cy.wait(300)
  })

  it('renders the keyword input field', () => {
    cy.get('input[type="text"], input[placeholder*="earch"], input[placeholder*="eyword"]')
      .should('exist')
    cy.wait(400)
  })

  it('can type a keyword into the search field', () => {
    cy.get('input[type="text"], input[placeholder*="earch"], input[placeholder*="eyword"]')
      .first()
      .type('phone')
    cy.wait(500)
    cy.get('input').first().should('have.value', 'phone')
  })

  it('shows Preview Listing button for mock listing', () => {
    cy.contains('button', 'Preview Listing').should('be.visible')
    cy.wait(400)
  })

  it('can open listing details from search results', () => {
    cy.contains('button', 'Preview Listing').click()
    cy.wait(600)
    cy.get('.search-modal.listing-details-modal', { timeout: 8000 }).should('exist')
    cy.get('.back-link-btn').first().click({ force: true })
    cy.wait(400)
  })

  it('has a Back button that returns to profile', () => {
    cy.get('.back-link-btn').click({ force: true })
    cy.wait(400)
    cy.contains('Welcome, Demo User').should('be.visible')
  })
})
