# Frontend Quick Start Guide

## Prerequisites
- Node.js (v18 or higher)
- npm or yarn package manager

## Setup Instructions

### 1. Install Dependencies
```bash
cd frontend
npm install
```

### 2. Environment Configuration
```bash
cp .env.example .env.local
```

Edit `.env.local` and set:
```
VITE_API_BASE_URL=http://localhost:8080
```

### 3. Run Development Server
```bash
npm run dev
```

The frontend will be available at: `http://localhost:5173`

## Available Scripts

### Development
```bash
npm run dev          # Start development server with hot reload
```

### Type Checking
```bash
npm run typecheck    # Run TypeScript type checking
```

### Linting
```bash
npm run lint         # Run ESLint to check code quality
```

### Build
```bash
npm run build        # Create production build
npm run preview      # Preview production build locally
```

## Testing the Password Toggle Feature

### Manual Testing Steps

1. **Start the application**
   ```bash
   npm run dev
   ```

2. **Test Login Form**
   - Open `http://localhost:5173`
   - You should see the login form by default
   - Look for the eye icon (👁️🗨️) next to the password field
   - Click the icon - password should become visible
   - Click again - password should be hidden
   - Type some text and verify the toggle works

3. **Test Register Form**
   - Click the "Register" tab
   - Find the password field with eye icon
   - Test the toggle functionality
   - Verify it works independently from the login form

4. **Test Keyboard Navigation**
   - Use Tab key to navigate to the password field
   - Tab again to reach the eye icon button
   - Press Enter or Space to toggle visibility
   - Verify focus indicators are visible

5. **Test Accessibility**
   - Use a screen reader (if available)
   - Verify ARIA labels are announced correctly
   - Check that the button purpose is clear

## Common Issues and Solutions

### Issue: npm command not found
**Solution:** Install Node.js from https://nodejs.org/

### Issue: Port 5173 already in use
**Solution:** 
```bash
# Kill the process using the port
lsof -ti:5173 | xargs kill -9
# Or use a different port
npm run dev -- --port 3000
```

### Issue: Backend connection errors
**Solution:** 
- Ensure backend is running on `http://localhost:8080`
- Check CORS settings in backend
- Verify `.env.local` has correct API URL

### Issue: TypeScript errors
**Solution:**
```bash
npm run typecheck  # Check for type errors
```

## Project Structure

```
frontend/
├── src/
│   ├── api/
│   │   └── auth.ts              # API calls for authentication
│   ├── components/
│   │   ├── CitySelector.tsx     # City selection component
│   │   ├── ForgotPassword.tsx   # Password reset UI
│   │   └── ProfileEdit.tsx      # Profile editing form
│   ├── types/
│   │   └── user.ts              # TypeScript type definitions
│   ├── App.tsx                  # Main application component (UPDATED)
│   ├── main.tsx                 # Application entry point
│   └── styles.css               # Global styles (UPDATED)
├── .env.example                 # Environment variables template
├── .env.local                   # Local environment config (create this)
├── package.json                 # Dependencies and scripts
├── tsconfig.json                # TypeScript configuration
└── vite.config.ts               # Vite build configuration
```

## Key Features Implemented

### ✅ Authentication
- User registration with validation
- User login with JWT tokens
- Session persistence via localStorage
- Logout functionality

### ✅ Password Management
- **NEW:** Password visibility toggle with eye icon
- Password strength validation (min 8 characters)
- Forgot password UI (backend integration pending)

### ✅ User Profile
- View user profile
- Edit profile (name, city, bio, photo)
- City selection with popular cities dropdown

### ✅ UI/UX
- Responsive design
- Modern gradient backgrounds
- Smooth animations
- Loading states
- Error handling and display

## Recording Demo Video

### For Sprint 1 Submission

1. **Start screen recording** (QuickTime, OBS, or built-in screen recorder)

2. **Demo Script:**
   - Show the login page
   - Demonstrate password toggle on login form
   - Switch to register tab
   - Demonstrate password toggle on register form
   - Show keyboard navigation (Tab + Enter)
   - Register a new account (or login if backend is running)
   - Show the profile page
   - Demonstrate city selection
   - Show profile editing

3. **Narration Points:**
   - "This is the ReSellution login page"
   - "We've added a password visibility toggle for better UX"
   - "Users can click the eye icon to show/hide their password"
   - "The feature works on both login and register forms"
   - "It's fully keyboard accessible and screen-reader friendly"

4. **Save and upload** the video for submission

## Next Steps for Sprint 2

Potential features to implement:
- [ ] Password strength indicator
- [ ] Confirm password field
- [ ] Email verification flow
- [ ] Social login integration
- [ ] Remember me checkbox
- [ ] Two-factor authentication UI

## Support

If you encounter any issues:
1. Check the browser console for errors
2. Verify backend is running
3. Check network tab for API call failures
4. Review the README.md in the root directory
5. Contact team lead or backend engineers

## Git Workflow

### Before making changes:
```bash
git pull origin main
```

### After making changes:
```bash
git add .
git commit -m "feat: add password visibility toggle to login/register forms"
git push origin main
```

### For feature branches:
```bash
git checkout -b feature/password-toggle
# Make changes
git add .
git commit -m "feat: add password visibility toggle"
git push origin feature/password-toggle
# Create pull request on GitHub
```

## Resources

- [React Documentation](https://react.dev/)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [Vite Guide](https://vitejs.dev/guide/)
- [MDN Web Docs](https://developer.mozilla.org/)
