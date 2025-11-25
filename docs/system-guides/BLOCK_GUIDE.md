# 🚫 Block User System Guide

## Overview

The SkillMatch platform includes a **user blocking system** that allows users to block other users for privacy, safety, and comfort. Blocked users cannot:
- Send messages to the blocker
- View the blocker's full profile
- Create bookings with the blocker
- See the blocker in search results (optional implementation)

**Key Features:**
- Block/unblock users
- View blocked users list
- Check block status
- Optional reason for blocking
- Bidirectional block checking

---

## Database Schema

### `blocks` Table

```sql
CREATE TABLE blocks (
    id SERIAL PRIMARY KEY,
    blocker_id INTEGER REFERENCES users(user_id) ON DELETE CASCADE,
    blocked_id INTEGER REFERENCES users(user_id) ON DELETE CASCADE,
    reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_not_self_block CHECK (blocker_id != blocked_id),
    UNIQUE(blocker_id, blocked_id)
);
```

**Key Points:**
- Prevents self-blocking via CHECK constraint
- One block per user pair
- Cascading delete when user is deleted
- Optional reason field

---

## API Endpoints

### 1. Block a User

**POST** `/blocks`

Blocks a user.

**Request Body:**
```json
{
  "blocked_user_id": 42,
  "reason": "Inappropriate messages"
}
```

**Validation:**
- Cannot block yourself
- Cannot block same user twice

**Response (Success):**
```json
{
  "message": "User blocked successfully",
  "block_id": 15
}
```

**Response (Error):**
```json
{
  "error": "User already blocked"
}
```

---

### 2. Unblock a User

**DELETE** `/blocks/:userId`

Removes a block.

**Response:**
```json
{
  "message": "User unblocked successfully"
}
```

---

### 3. Get Blocked Users List

**GET** `/blocks`

Returns all users you have blocked.

**Response:**
```json
{
  "blocked_users": [
    {
      "user_id": 42,
      "username": "john_doe",
      "profile_image_url": "https://...",
      "google_profile_picture": null,
      "reason": "Inappropriate messages",
      "blocked_at": "2025-11-13T10:00:00Z"
    }
  ],
  "total": 1
}
```

---

### 4. Check Block Status

**GET** `/blocks/check/:userId`

Check if you have blocked a user or if they have blocked you.

**Response:**
```json
{
  "is_blocked": true,      // You blocked this user
  "is_blocked_by": false   // This user blocked you
}
```

**Use Cases:**
- Hide "Message" button if blocked
- Show "Unblock" button if already blocked
- Display "This user has blocked you" message

---

## Frontend Implementation

### React Block User Component

```typescript
import React, { useState, useEffect } from 'react';
import axios from 'axios';

interface BlockUserModalProps {
  userId: number;
  userName: string;
  onClose: () => void;
  onBlocked?: () => void;
}

export const BlockUserModal: React.FC<BlockUserModalProps> = ({ 
  userId, 
  userName, 
  onClose,
  onBlocked 
}) => {
  const [reason, setReason] = useState('');
  const [loading, setLoading] = useState(false);

  const handleBlock = async () => {
    setLoading(true);
    try {
      await axios.post('/blocks', {
        blocked_user_id: userId,
        reason: reason || undefined
      });
      alert(`${userName} has been blocked`);
      onBlocked?.();
      onClose();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to block user');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="modal">
      <div className="modal-content">
        <h2>Block {userName}?</h2>
        
        <p>Blocked users cannot:</p>
        <ul>
          <li>Send you messages</li>
          <li>Book your services</li>
          <li>View your full profile</li>
        </ul>

        <div className="form-group">
          <label>Reason (Optional)</label>
          <textarea
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder="Why are you blocking this user?"
            rows={3}
          />
        </div>

        <div className="actions">
          <button 
            onClick={handleBlock} 
            disabled={loading}
            className="btn-danger"
          >
            {loading ? 'Blocking...' : 'Block User'}
          </button>
          <button onClick={onClose}>Cancel</button>
        </div>
      </div>
    </div>
  );
};

// Blocked Users List Component
export const BlockedUsersList: React.FC = () => {
  const [blockedUsers, setBlockedUsers] = useState<any[]>([]);

  useEffect(() => {
    fetchBlockedUsers();
  }, []);

  const fetchBlockedUsers = async () => {
    const response = await axios.get('/blocks');
    setBlockedUsers(response.data.blocked_users);
  };

  const handleUnblock = async (userId: number, userName: string) => {
    if (confirm(`Unblock ${userName}?`)) {
      await axios.delete(`/blocks/${userId}`);
      fetchBlockedUsers();
    }
  };

  return (
    <div className="blocked-users-list">
      <h2>Blocked Users</h2>
      
      {blockedUsers.length === 0 ? (
        <p>You haven't blocked anyone</p>
      ) : (
        <div className="user-list">
          {blockedUsers.map(user => (
            <div key={user.user_id} className="user-item">
              <img 
                src={user.profile_image_url || user.google_profile_picture || '/default-avatar.png'} 
                alt={user.username}
              />
              <div className="user-info">
                <h3>{user.username}</h3>
                {user.reason && <p className="reason">Reason: {user.reason}</p>}
                <small>Blocked on {new Date(user.blocked_at).toLocaleDateString()}</small>
              </div>
              <button 
                onClick={() => handleUnblock(user.user_id, user.username)}
                className="btn-secondary"
              >
                Unblock
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
```

---

## Integration with Other Features

### 1. Messaging System

**Prevent blocked users from sending messages:**

```go
// In message_handlers.go - SendMessage function
func SendMessage(c *gin.Context) {
    senderID := c.GetInt("userID")
    
    var input struct {
        ConversationID int    `json:"conversation_id"`
        Content        string `json:"content"`
    }
    c.ShouldBindJSON(&input)
    
    // Get receiver from conversation
    var receiverID int
    // ... query to get receiverID
    
    // Check if blocked
    if IsUserBlocked(db, ctx, receiverID, senderID) {
        c.JSON(403, gin.H{"error": "You cannot message this user"})
        return
    }
    
    if IsUserBlocked(db, ctx, senderID, receiverID) {
        c.JSON(403, gin.H{"error": "This user has blocked you"})
        return
    }
    
    // Continue with message sending...
}
```

### 2. Booking System

**Prevent blocked users from booking:**

```go
// In booking_handlers.go - CreateBooking function
func createBookingHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
    return func(c *gin.Context) {
        clientID := c.GetInt("userID")
        providerID := // ... from input
        
        // Check if blocked
        if IsUserBlocked(dbPool, ctx, providerID, clientID) {
            c.JSON(403, gin.H{"error": "You cannot book with this provider"})
            return
        }
        
        if IsUserBlocked(dbPool, ctx, clientID, providerID) {
            c.JSON(403, gin.H{"error": "This provider has blocked you"})
            return
        }
        
        // Continue with booking...
    }
}
```

### 3. Browse/Search

**Filter out blocked users from search results:**

```sql
SELECT * FROM users
WHERE user_id NOT IN (
    -- Users I blocked
    SELECT blocked_id FROM blocks WHERE blocker_id = $1
    UNION
    -- Users who blocked me
    SELECT blocker_id FROM blocks WHERE blocked_id = $1
)
```

### 4. Profile Viewing

**Show limited profile for blocked users:**

```typescript
useEffect(() => {
  const checkBlockStatus = async () => {
    const response = await axios.get(`/blocks/check/${providerId}`);
    
    if (response.data.is_blocked_by) {
      // Provider blocked you - show limited profile
      setProfileLimited(true);
    }
    
    if (response.data.is_blocked) {
      // You blocked this provider
      setShowUnblockButton(true);
    }
  };
  
  checkBlockStatus();
}, [providerId]);
```

---

## Helper Function

The system includes a helper function for easy block checking:

```go
// In block_handlers.go
func IsUserBlocked(dbPool *pgxpool.Pool, ctx context.Context, blockerID, blockedID int) bool {
    var exists bool
    err := dbPool.QueryRow(ctx, `
        SELECT EXISTS(SELECT 1 FROM blocks WHERE blocker_id = $1 AND blocked_id = $2)
    `, blockerID, blockedID).Scan(&exists)
    
    return err == nil && exists
}
```

**Usage in other handlers:**

```go
if IsUserBlocked(dbPool, ctx, userA, userB) {
    // userA has blocked userB
    return
}
```

---

## User Experience Guidelines

### When to Show Block Option
- On user profiles
- In message conversations
- After bookings (if negative experience)
- In search results (three-dot menu)

### UI/UX Best Practices

1. **Confirmation Dialog**
   - Always ask for confirmation before blocking
   - Explain consequences clearly
   - Optional reason field (helps with platform moderation)

2. **Visual Indicators**
   - Show "Blocked" badge in blocked users list
   - Disable messaging/booking buttons when blocked
   - Display "This user has blocked you" message

3. **Easy Unblocking**
   - Provide easy access to blocked users list
   - One-click unblock with confirmation
   - Show when user was blocked

4. **Privacy**
   - Don't notify blocked user
   - Don't show who blocked whom publicly
   - Blocked users just see "unavailable" or "not found"

---

## Security Considerations

### 1. Prevent Abuse
- Rate limit block actions (max 10 per hour)
- Log block actions for admin review
- Flag users with excessive blocks

### 2. Data Privacy
- Don't expose block reasons to blocked users
- Admin-only access to block statistics
- GDPR compliance: allow block data export/deletion

### 3. Cascading Effects
- When user is deleted, all their blocks are removed (CASCADE)
- Blocked users are automatically unfavorited
- Clear any pending bookings

---

## Admin Monitoring

### Get Block Statistics

```sql
-- Most blocked users
SELECT 
    blocked_id,
    u.username,
    COUNT(*) as times_blocked
FROM blocks b
JOIN users u ON b.blocked_id = u.user_id
GROUP BY blocked_id, u.username
ORDER BY times_blocked DESC
LIMIT 20;

-- Users who block most frequently
SELECT 
    blocker_id,
    u.username,
    COUNT(*) as blocks_made
FROM blocks b
JOIN users u ON b.blocker_id = u.user_id
GROUP BY blocker_id, u.username
ORDER BY blocks_made DESC
LIMIT 20;

-- Block reasons distribution
SELECT 
    reason,
    COUNT(*) as count
FROM blocks
WHERE reason IS NOT NULL
GROUP BY reason
ORDER BY count DESC;
```

---

## Testing

### Manual Testing

**1. Block a user:**
```bash
curl -X POST http://localhost:8080/blocks \
  -H "Authorization: Bearer USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "blocked_user_id": 42,
    "reason": "Spam messages"
  }'
```

**2. Check block status:**
```bash
curl -X GET http://localhost:8080/blocks/check/42 \
  -H "Authorization: Bearer USER_TOKEN"
```

**3. Get blocked users:**
```bash
curl -X GET http://localhost:8080/blocks \
  -H "Authorization: Bearer USER_TOKEN"
```

**4. Unblock a user:**
```bash
curl -X DELETE http://localhost:8080/blocks/42 \
  -H "Authorization: Bearer USER_TOKEN"
```

---

## Future Enhancements

- [ ] Temporary blocks (auto-unblock after X days)
- [ ] Report user when blocking
- [ ] Block entire account network (multiple accounts)
- [ ] AI-powered block suggestions
- [ ] Mutual block detection
- [ ] Block appeals system
- [ ] Export blocked users list

---

**Last Updated:** November 13, 2025
