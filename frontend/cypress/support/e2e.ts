// Cypress support file — add custom commands here

// Slow down between each test so steps are visible in the UI runner
afterEach(() => {
  cy.wait(800)
})

// Helper: enter preview mode (reusable across specs)
Cypress.Commands.add('enterPreview', () => {
  cy.visit('/')
  cy.contains('Preview app without login').click()
  cy.contains('Welcome, Demo User').should('be.visible')
})

declare global {
  namespace Cypress {
    interface Chainable {
      enterPreview(): Chainable<void>
    }
  }
}
