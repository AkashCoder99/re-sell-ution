/**
 * Cypress E2E — Login form interaction
 * Verifies the login form renders, accepts input, and shows feedback.
 */
describe('Login form', () => {
  beforeEach(() => {
    cy.visit('/')
  })

  it('renders the ReSellution heading', () => {
    cy.contains('ReSellution').should('be.visible')
  })

  it('renders the Login tab as active by default', () => {
    cy.get('.auth-mode-btn.active').should('contain.text', 'Login')
  })

  it('fills in email and password fields', () => {
    cy.get('#login-email').type('test@example.com')
    cy.get('#login-password').type('Password1')
    cy.get('#login-email').should('have.value', 'test@example.com')
    cy.get('#login-password').should('have.value', 'Password1')
  })

  it('toggles password visibility', () => {
    cy.get('#login-password').should('have.attr', 'type', 'password')
    cy.get('.auth-password-toggle').first().click()
    cy.get('#login-password').should('have.attr', 'type', 'text')
  })

  it('switches to Register tab', () => {
    cy.contains('button', 'Register').click()
    cy.get('.auth-mode-btn.active').should('contain.text', 'Register')
    cy.get('#register-name').should('exist')
  })

  it('fills in the register form', () => {
    cy.contains('button', 'Register').click()
    cy.get('#register-name').type('Krishna')
    cy.get('#register-email').type('krishna@example.com')
    cy.get('#register-password').type('Password1')
    cy.get('#register-name').should('have.value', 'Krishna')
    cy.get('#register-email').should('have.value', 'krishna@example.com')
  })

  it('shows forgot password link and navigates to it', () => {
    cy.contains('Forgot password?').click()
    cy.contains(/forgot/i).should('be.visible')
  })

  it('shows preview mode button and enters preview', () => {
    cy.contains('Preview app without login').click()
    cy.contains('Demo User').should('be.visible')
  })
})
