describe('Chat and inbox flow', () => {
  function enterPreviewMode() {
    cy.visit('/')
    cy.contains('Preview app without login').click()
    cy.contains('Welcome, Demo User').should('be.visible')
  }

  function startChatFromPreviewListing(message: string) {
    cy.contains('button', 'Search by Keyword').click()
    cy.wait(600)
    cy.contains('button', 'Preview Listing').click()
    cy.wait(600)
    cy.contains('button', 'Chat with Seller').scrollIntoView().click({ force: true })
    cy.wait(800)
    cy.get('.chat-thread', { timeout: 10000 }).should('be.visible')
    cy.get('.chat-thread input[placeholder="Type your message..."]').type(message)
    cy.wait(500)
    cy.contains('.chat-thread button', 'Send').click()
    cy.wait(500)
  }

  it('starts chat from browse and sends a message', () => {
    enterPreviewMode()

    startChatFromPreviewListing('Hi, is this still available?')
    cy.contains('.chat-thread-bubble', 'Hi, is this still available?').should('be.visible')
  })

  it('shows inbox entry after starting a chat', () => {
    enterPreviewMode()

    startChatFromPreviewListing('Checking inbox visibility')
    cy.get('.chat-thread .back-link-btn').click()

    cy.get('body').then(($body) => {
      if ($body.find('button:contains("Inbox")').length === 0) {
        cy.contains('button', 'Back to Profile').click()
      }
    })

    cy.contains('button', 'Inbox').click()
    cy.contains('h2', 'Inbox').should('be.visible')
    cy.get('.conversation-inbox-list li').should('have.length.greaterThan', 0)
    cy.get('.conversation-inbox-pagination').should('contain.text', 'Page 1')
  })
})
