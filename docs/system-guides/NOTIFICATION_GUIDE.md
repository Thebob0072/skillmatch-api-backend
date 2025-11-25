# 🔔 Notification System Guide

## Overview

The SkillMatch platform includes a **comprehensive notification system** that keeps users informed about important events in real-time. The system supports:

- Real-time notifications via WebSocket
- Multiple notification types
- Read/unread status tracking
- Notification history
- Bulk operations (mark all as read, delete all)

---

## Database Schema

### `notifications` Table

```sql
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(user_id),
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    metadata JSONB,  -- Additional context data
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_type CHECK (type IN (
        'new_message', 
        'booking_request', 
        'booking_confirmed', 
        'booking_cancelled', 
        'booking_completed',
        'kyc_approved', 
        'kyc_rejected', 
        'new_review',
        'payment_success',
        'payment_failed',
        'tier_upgraded'
    ))
);
```

**Key Points:**
- `metadata` stores additional JSON data (booking_id, conversation_id, etc.)
- Tracks read status and timestamp
- Type-checked for valid notification types

---

## Notification Types

| Type | Description | Metadata |
|------|-------------|----------|
| `new_message` | New chat message received | `conversation_id`, `sender_id`, `message_id` |
| `booking_request` | New booking request received (provider) | `booking_id`, `client_id` |
| `booking_confirmed` | Booking confirmed by provider | `booking_id`, `provider_id` |
| `booking_cancelled` | Booking cancelled | `booking_id`, `reason` |
| `booking_completed` | Booking marked as completed | `booking_id` |
| `kyc_approved` | KYC verification approved | `verification_id` |
| `kyc_rejected` | KYC verification rejected | `verification_id`, `reason` |
| `new_review` | New review received | `review_id`, `client_id`, `rating` |
| `payment_success` | Payment completed successfully | `payment_intent_id`, `amount` |
| `payment_failed` | Payment failed | `payment_intent_id`, `error` |
| `tier_upgraded` | Subscription tier upgraded | `old_tier`, `new_tier` |

---

## API Endpoints

### 1. Get All Notifications

**GET** `/notifications?limit=50&offset=0&type=new_message`

Returns all notifications for the authenticated user.

**Query Parameters:**
- `limit` (optional): Number of notifications (default: 50, max: 100)
- `offset` (optional): Pagination offset (default: 0)
- `type` (optional): Filter by notification type

**Response:**
```json
{
  "notifications": [
    {
      "id": 10,
      "user_id": 5,
      "type": "booking_request",
      "title": "New Booking Request",
      "message": "You have a new booking request",
      "metadata": {
        "booking_id": 42,
        "client_id": 8
      },
      "is_read": false,
      "created_at": "2025-11-13T10:00:00Z"
    },
    {
      "id": 9,
      "user_id": 5,
      "type": "new_message",
      "title": "New Message",
      "message": "You have a new message",
      "metadata": {
        "conversation_id": 3,
        "sender_id": 7,
        "message_id": 156
      },
      "is_read": true,
      "read_at": "2025-11-13T09:30:00Z",
      "created_at": "2025-11-13T09:15:00Z"
    }
  ],
  "total": 25,
  "unread_count": 5
}
```

---

### 2. Get Unread Notification Count

**GET** `/notifications/unread/count`

Returns the count of unread notifications.

**Response:**
```json
{
  "unread_count": 5
}
```

**Use Cases:**
- Display badge count on notification bell icon
- Update UI in real-time

---

### 3. Mark Notification as Read

**PATCH** `/notifications/:id/read`

Marks a single notification as read.

**Response:**
```json
{
  "message": "Notification marked as read"
}
```

---

### 4. Mark All Notifications as Read

**PATCH** `/notifications/read-all`

Marks all notifications as read for the current user.

**Response:**
```json
{
  "message": "All notifications marked as read",
  "updated_count": 12
}
```

---

### 5. Delete a Notification

**DELETE** `/notifications/:id`

Deletes a specific notification.

**Response:**
```json
{
  "message": "Notification deleted"
}
```

---

### 6. Delete All Notifications

**DELETE** `/notifications`

Deletes all notifications for the current user.

**Response:**
```json
{
  "message": "All notifications deleted",
  "deleted_count": 25
}
```

---

## Creating Notifications (Backend)

### Internal Function

The `CreateNotification()` function is used internally to create notifications:

```go
func CreateNotification(userID int, notifType, message string, metadata map[string]interface{}) error
```

**Example Usage:**

```go
// When a new booking is created
CreateNotification(providerID, "booking_request", "You have a new booking request", map[string]interface{}{
    "booking_id": bookingID,
    "client_id":  clientID,
})

// When a message is sent
CreateNotification(receiverID, "new_message", "You have a new message", map[string]interface{}{
    "conversation_id": conversationID,
    "sender_id":       senderID,
    "message_id":      messageID,
})

// When KYC is approved
CreateNotification(userID, "kyc_approved", "Your KYC verification has been approved!", map[string]interface{}{
    "verification_id": verificationID,
})
```

---

## WebSocket Real-Time Notifications

Notifications are automatically sent via WebSocket when created.

**Received Message Format:**

```json
{
  "type": "notification",
  "payload": {
    "id": 15,
    "type": "booking_confirmed",
    "title": "Booking Confirmed",
    "message": "Your booking has been confirmed by the provider",
    "metadata": {
      "booking_id": 42,
      "provider_id": 5
    }
  }
}
```

---

## Frontend Implementation

### React Notification Component

```typescript
import React, { useState, useEffect } from 'react';
import axios from 'axios';

interface Notification {
  id: number;
  type: string;
  title: string;
  message: string;
  metadata?: any;
  is_read: boolean;
  created_at: string;
}

export const NotificationBell: React.FC = () => {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [isOpen, setIsOpen] = useState(false);

  // Fetch notifications
  useEffect(() => {
    fetchNotifications();
    fetchUnreadCount();
  }, []);

  // WebSocket listener
  useEffect(() => {
    const ws = new WebSocket('ws://localhost:8080/ws');
    
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      
      if (data.type === 'notification') {
        setNotifications(prev => [data.payload, ...prev]);
        setUnreadCount(prev => prev + 1);
        
        // Show browser notification
        if (Notification.permission === 'granted') {
          new Notification(data.payload.title, {
            body: data.payload.message,
            icon: '/logo.png'
          });
        }
      }
    };
    
    return () => ws.close();
  }, []);

  const fetchNotifications = async () => {
    const response = await axios.get('/notifications');
    setNotifications(response.data.notifications);
    setUnreadCount(response.data.unread_count);
  };

  const fetchUnreadCount = async () => {
    const response = await axios.get('/notifications/unread/count');
    setUnreadCount(response.data.unread_count);
  };

  const markAsRead = async (id: number) => {
    await axios.patch(`/notifications/${id}/read`);
    setNotifications(prev =>
      prev.map(n => n.id === id ? { ...n, is_read: true } : n)
    );
    setUnreadCount(prev => Math.max(0, prev - 1));
  };

  const markAllAsRead = async () => {
    await axios.patch('/notifications/read-all');
    setNotifications(prev => prev.map(n => ({ ...n, is_read: true })));
    setUnreadCount(0);
  };

  const handleNotificationClick = (notif: Notification) => {
    markAsRead(notif.id);
    
    // Navigate based on notification type
    switch (notif.type) {
      case 'new_message':
        window.location.href = `/chat/${notif.metadata.conversation_id}`;
        break;
      case 'booking_request':
      case 'booking_confirmed':
        window.location.href = `/bookings/${notif.metadata.booking_id}`;
        break;
      case 'new_review':
        window.location.href = `/reviews`;
        break;
    }
  };

  return (
    <div className="notification-bell">
      <button onClick={() => setIsOpen(!isOpen)} className="bell-icon">
        🔔
        {unreadCount > 0 && (
          <span className="badge">{unreadCount}</span>
        )}
      </button>

      {isOpen && (
        <div className="notification-dropdown">
          <div className="header">
            <h3>Notifications</h3>
            {unreadCount > 0 && (
              <button onClick={markAllAsRead}>Mark all as read</button>
            )}
          </div>

          <div className="notification-list">
            {notifications.length === 0 ? (
              <p className="empty">No notifications</p>
            ) : (
              notifications.map(notif => (
                <div
                  key={notif.id}
                  className={`notification-item ${notif.is_read ? 'read' : 'unread'}`}
                  onClick={() => handleNotificationClick(notif)}
                >
                  <div className="icon">{getNotificationIcon(notif.type)}</div>
                  <div className="content">
                    <h4>{notif.title}</h4>
                    <p>{notif.message}</p>
                    <small>{formatTime(notif.created_at)}</small>
                  </div>
                  {!notif.is_read && <div className="unread-dot" />}
                </div>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
};

function getNotificationIcon(type: string): string {
  const icons: Record<string, string> = {
    new_message: '💬',
    booking_request: '📅',
    booking_confirmed: '✅',
    booking_cancelled: '❌',
    booking_completed: '🎉',
    kyc_approved: '✅',
    kyc_rejected: '❌',
    new_review: '⭐',
    payment_success: '💳',
    payment_failed: '⚠️',
    tier_upgraded: '🚀',
  };
  return icons[type] || '🔔';
}

function formatTime(timestamp: string): string {
  const date = new Date(timestamp);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  
  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffMins < 1440) return `${Math.floor(diffMins / 60)}h ago`;
  return date.toLocaleDateString();
}
```

---

## Browser Push Notifications

### Request Permission

```javascript
// Request permission on app load
if ('Notification' in window && Notification.permission === 'default') {
  Notification.requestPermission();
}
```

### Show Notification

```javascript
function showBrowserNotification(title: string, message: string) {
  if (Notification.permission === 'granted') {
    const notification = new Notification(title, {
      body: message,
      icon: '/logo.png',
      badge: '/badge.png',
      tag: 'skillmatch-notification',
      requireInteraction: false
    });

    notification.onclick = () => {
      window.focus();
      notification.close();
    };
  }
}
```

---

## Integration Points

### 1. Booking System

**When booking is created:**
```go
CreateNotification(providerID, "booking_request", "You have a new booking request", map[string]interface{}{
    "booking_id": bookingID,
    "client_id":  clientID,
})
```

**When booking is confirmed:**
```go
CreateNotification(clientID, "booking_confirmed", "Your booking has been confirmed", map[string]interface{}{
    "booking_id":   bookingID,
    "provider_id":  providerID,
})
```

**When booking is cancelled:**
```go
CreateNotification(otherUserID, "booking_cancelled", "A booking has been cancelled", map[string]interface{}{
    "booking_id": bookingID,
    "reason":     cancellationReason,
})
```

### 2. Messaging System

**When message is sent:**
```go
CreateNotification(receiverID, "new_message", "You have a new message", map[string]interface{}{
    "conversation_id": conversationID,
    "sender_id":       senderID,
    "message_id":      messageID,
})
```

### 3. KYC Verification

**When KYC is approved:**
```go
CreateNotification(userID, "kyc_approved", "Your KYC verification has been approved!", nil)
```

**When KYC is rejected:**
```go
CreateNotification(userID, "kyc_rejected", "Your KYC verification was rejected", map[string]interface{}{
    "reason": rejectionReason,
})
```

### 4. Review System

**When provider receives a review:**
```go
CreateNotification(providerID, "new_review", "You received a new review!", map[string]interface{}{
    "review_id": reviewID,
    "client_id": clientID,
    "rating":    rating,
})
```

### 5. Payment System

**When payment succeeds:**
```go
CreateNotification(userID, "payment_success", "Payment completed successfully", map[string]interface{}{
    "payment_intent_id": paymentIntentID,
    "amount":            amount,
})
```

---

## Best Practices

### 1. Notification Frequency
- Don't overwhelm users with too many notifications
- Group similar notifications (e.g., "3 new messages" instead of 3 separate notifications)
- Allow users to customize notification preferences

### 2. Actionable Notifications
- Include relevant data in `metadata` for navigation
- Provide clear call-to-action
- Link directly to relevant content

### 3. Timing
- Send notifications immediately for urgent items (bookings, messages)
- Batch non-urgent notifications (reviews, tips)
- Respect user's timezone

### 4. Cleanup
- Periodically delete old read notifications (e.g., after 30 days)
- Implement auto-delete for low-priority notifications
- Allow users to delete all notifications

---

## Testing

### Manual Testing

**1. Create a notification:**
```go
// In Go code
CreateNotification(5, "booking_request", "Test notification", map[string]interface{}{
    "booking_id": 1,
})
```

**2. Get notifications:**
```bash
curl -X GET http://localhost:8080/notifications \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**3. Mark as read:**
```bash
curl -X PATCH http://localhost:8080/notifications/1/read \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## Troubleshooting

### Notifications Not Appearing

**Symptom:** Notifications created but not appearing in list.

**Solutions:**
1. Check `user_id` matches the authenticated user
2. Verify notification type is valid
3. Check database for created notification
4. Refresh notification list

### WebSocket Notifications Not Received

**Symptom:** Notifications appear after refresh but not in real-time.

**Solutions:**
1. Verify WebSocket connection is active
2. Check `wsManager.BroadcastToUser()` is called
3. Ensure user is connected to WebSocket
4. Check browser console for WebSocket errors

---

## Future Enhancements

- [ ] Email notifications
- [ ] SMS notifications
- [ ] Push notifications (mobile apps)
- [ ] Notification preferences/settings
- [ ] Notification categories (mute certain types)
- [ ] Scheduled notifications
- [ ] Notification templates
- [ ] Multi-language notification messages

---

**Last Updated:** November 13, 2025
