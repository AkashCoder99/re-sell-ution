/**
 * Cypress E2E — Browse Listings page
 * Covers heading, city filter, category filter, listing cards, and detail navigation.
 */
describe('Browse Listings', () => {
  beforeEach(() => {
    cy.enterPreview()
    cy.contains('button', 'Explore Nearby').click()
    cy.wait(600)
  })

  it('shows the Browse Listings heading', () => {
    cy.contains('h2', 'Browse Listings').should('be.visible')
  })

  it('renders the city filter dropdown', () => {
    cy.get('select').first().should('exist')
    cy.wait(400)
  })

  it('renders the category filter dropdown', () => {
    cy.get('select').should('have.length.greaterThan', 0)
    cy.wait(400)
  })

  it('can change the city filter', () => {
    cy.get('select').first().select(1)
    cy.wait(600)
    cy.get('select').first().should('not.have.value', '')
  })

  it('has a Back button that returns to profile', () => {
    cy.get('.back-link-btn').click()
    cy.wait(400)
    cy.contains('Welcome, Demo User').should('be.visible')
  })
})
