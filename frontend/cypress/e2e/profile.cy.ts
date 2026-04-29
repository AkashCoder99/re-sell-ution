/**
 * Cypress E2E — Profile page
 * Covers profile display, navigation to sub-views, and logout.
 */
describe('Profile page', () => {
  beforeEach(() => {
    cy.enterPreview()
    cy.wait(400)
  })

  it('shows the welcome message with Demo User', () => {
    cy.contains('Welcome, Demo User').should('be.visible')
  })

  it('shows the Explore Nearby button', () => {
    cy.contains('button', 'Explore Nearby').should('be.visible')
    cy.wait(300)
  })

  it('shows the Search by Keyword button', () => {
    cy.contains('button', 'Search by Keyword').should('be.visible')
    cy.wait(300)
  })

  it('shows the Create Listing button', () => {
    cy.contains('button', 'Create Listing').should('be.visible')
    cy.wait(300)
  })

  it('navigates to Create Listing and shows step 1', () => {
    cy.contains('button', 'Create Listing').click()
    cy.wait(600)
    cy.get('h2.create-listing-heading').should('be.visible')
    cy.contains('button', 'Cancel').click({ force: true })
    cy.wait(400)
    cy.contains('Welcome, Demo User').should('be.visible')
  })

  it('navigates to My Listings', () => {
    cy.contains('button', 'My Listings').click()
    cy.wait(600)
    cy.contains('My Listings').should('be.visible')
    cy.get('.back-link-btn').click({ force: true })
    cy.wait(400)
    cy.contains('Welcome, Demo User').should('be.visible')
  })

  it('navigates to Search and back', () => {
    cy.contains('button', 'Search by Keyword').click()
    cy.wait(600)
    cy.get('h2.search-title').should('be.visible')
    cy.get('.back-link-btn').click({ force: true })
    cy.wait(400)
    cy.contains('Welcome, Demo User').should('be.visible')
  })

  it('can log out and returns to login screen', () => {
    cy.contains('button', 'Logout').click()
    cy.wait(600)
    cy.contains('ReSellution').should('be.visible')
    cy.get('.auth-mode-btn.active').should('contain.text', 'Login')
  })
})
