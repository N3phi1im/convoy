# User Stories

## Epic 1: User Management

### US-1.1: User Registration
**As a** new user  
**I want to** create an account with email and password  
**So that** I can access the platform and create routes

**Acceptance Criteria:**
- User can register with email, password, and display name
- Email must be unique and valid format
- Password must be at least 8 characters with complexity requirements
- User receives confirmation email (future enhancement)
- Account is created in the database

**Priority:** High

---

### US-1.2: User Login
**As a** registered user  
**I want to** log in with my credentials  
**So that** I can access my routes and profile

**Acceptance Criteria:**
- User can log in with email and password
- System returns JWT token on successful login
- Token expires after configured time period
- Failed login attempts are logged
- Clear error messages for invalid credentials

**Priority:** High

---

### US-1.3: User Profile
**As a** logged-in user  
**I want to** view and edit my profile  
**So that** I can manage my personal information

**Acceptance Criteria:**
- User can view their profile (name, email, join date)
- User can update display name and profile picture
- User can change password
- Email changes require verification (future)

**Priority:** Medium

---

## Epic 2: Route Creation

### US-2.1: Create Basic Route
**As a** logged-in user  
**I want to** create a route by selecting points on a map  
**So that** I can share my planned journey with others

**Acceptance Criteria:**
- User can add multiple waypoints by clicking on map
- System calculates route between waypoints using map provider
- User can name the route and add description
- User can set route visibility (public/private)
- Route is saved to database with creator information

**Priority:** High

---

### US-2.2: Route Details
**As a** route creator  
**I want to** add detailed information to my route  
**So that** participants know what to expect

**Acceptance Criteria:**
- User can set start date/time
- User can add route type (cycling, running, driving, walking)
- User can set difficulty level
- User can add notes and special instructions
- User can set maximum participants

**Priority:** High

---

### US-2.3: Edit Route
**As a** route creator  
**I want to** modify my route after creation  
**So that** I can adjust plans as needed

**Acceptance Criteria:**
- Creator can modify waypoints
- Creator can update route details
- Participants are notified of changes
- Route history is maintained
- Cannot edit routes that have already started

**Priority:** Medium

---

### US-2.4: Delete Route
**As a** route creator  
**I want to** delete a route I created  
**So that** I can remove cancelled plans

**Acceptance Criteria:**
- Only creator can delete route
- Participants are notified of deletion
- Soft delete with retention period
- Cannot delete routes in progress

**Priority:** Medium

---

## Epic 3: Route Discovery and Joining

### US-3.1: Browse Routes
**As a** logged-in user  
**I want to** browse available public routes  
**So that** I can find routes to join

**Acceptance Criteria:**
- User can view list of public routes
- Routes show basic info (name, date, type, participants)
- User can filter by date, type, location
- User can search by name or description
- Pagination for large result sets

**Priority:** High

---

### US-3.2: View Route Details
**As a** user  
**I want to** view detailed information about a route  
**So that** I can decide if I want to join

**Acceptance Criteria:**
- User can see full route on map
- Display all route details and waypoints
- Show list of current participants
- Show distance and estimated duration
- Display creator information

**Priority:** High

---

### US-3.3: Join Route
**As a** logged-in user  
**I want to** join an existing route  
**So that** I can participate in the journey

**Acceptance Criteria:**
- User can request to join public routes
- User is added to participant list
- Creator is notified of new participant
- User cannot join if route is full
- User cannot join routes that have started

**Priority:** High

---

### US-3.4: Leave Route
**As a** route participant  
**I want to** leave a route I joined  
**So that** I can cancel my participation

**Acceptance Criteria:**
- User can leave route before it starts
- Creator is notified when someone leaves
- User is removed from participant list
- User can rejoin if spots available

**Priority:** Medium

---

## Epic 4: Route Sharing

### US-4.1: Share Route Link
**As a** route creator  
**I want to** generate a shareable link for my route  
**So that** I can invite specific people

**Acceptance Criteria:**
- System generates unique shareable URL
- Link works for both logged-in and non-logged-in users
- Non-logged-in users prompted to register/login to join
- Link can be copied to clipboard
- Link shows route preview

**Priority:** High

---

### US-4.3: Route Privacy Settings
**As a** route creator  
**I want to** control who can see and join my route  
**So that** I can manage privacy

**Acceptance Criteria:**
- Routes can be public, private, or invite-only
- Public routes appear in browse/search
- Private routes only accessible via direct link
- Invite-only requires creator approval to join

**Priority:** Medium

---

## Epic 5: Participant Management

### US-5.1: View Participants
**As a** route member  
**I want to** see who else is joining  
**So that** I know who will be on the route

**Acceptance Criteria:**
- Display list of all participants
- Show participant names and profile pictures
- Indicate route creator
- Show join timestamp

**Priority:** Medium

---

### US-5.2: Remove Participant
**As a** route creator  
**I want to** remove participants from my route  
**So that** I can manage the group

**Acceptance Criteria:**
- Creator can remove any participant
- Removed user is notified
- Removed user cannot rejoin without permission
- Cannot remove self (use delete route instead)

**Priority:** Low

---

## Epic 6: Map Integration

### US-6.1: Interactive Map Display
**As a** user  
**I want to** interact with the map  
**So that** I can explore routes visually

**Acceptance Criteria:**
- Map displays with zoom/pan controls
- Route path clearly visible
- Waypoints marked with icons
- Map supports different base layers (street, satellite)
- Responsive on mobile devices

**Priority:** High

---

### US-6.2: Turn-by-Turn Directions
**As a** route participant  
**I want to** get turn-by-turn directions  
**So that** I can navigate the route

**Acceptance Criteria:**
- System provides step-by-step directions
- Directions include distance and time estimates
- Directions available in text format
- Export to GPS device/app (future)

**Priority:** Medium

---

### US-6.3: Route Statistics
**As a** user viewing a route  
**I want to** see route statistics  
**So that** I can understand the route difficulty

**Acceptance Criteria:**
- Display total distance
- Show estimated duration
- Display elevation gain/loss (if available)
- Show route type icon
- Calculate based on selected transport mode

**Priority:** Medium

---

## Epic 7: Real-Time Features (Future)

### US-7.1: Live Location Tracking
**As a** route participant  
**I want to** share my live location during the route  
**So that** others can see where I am

**Acceptance Criteria:**
- User can enable/disable location sharing
- Location updates in real-time on map
- Privacy controls for location visibility
- Location history for route duration
- Battery-efficient updates

**Priority:** Low

---

### US-7.2: Route Progress
**As a** route participant  
**I want to** see my progress along the route  
**So that** I know how far I've come

**Acceptance Criteria:**
- Display percentage complete
- Show distance remaining
- Estimate time to completion
- Highlight completed waypoints

**Priority:** Low

---

## Summary

### MVP (Phase 1)
- US-1.1, US-1.2: Authentication
- US-2.1, US-2.2: Route creation
- US-3.1, US-3.2, US-3.3: Route discovery and joining
- US-4.1: Route sharing
- US-6.1: Map display

### Phase 2
- US-1.3: User profiles
- US-2.3, US-2.4: Route management
- US-3.4: Leave routes
- US-4.3: Privacy settings
- US-5.1: View participants
- US-6.2, US-6.3: Enhanced map features

### Phase 3 (Future)
- US-5.2, US-5.3: Participant management
- US-7.1, US-7.2: Real-time tracking

