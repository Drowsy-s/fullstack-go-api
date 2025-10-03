import React, { useEffect, useState } from 'react';
import { updateUser } from '../api';
import { useAuth } from '../context/AuthContext';

const ProfilePage: React.FC = () => {
  const { user, token, refreshProfile } = useAuth();
  const [name, setName] = useState(user?.name ?? '');
  const [email, setEmail] = useState(user?.email ?? '');
  const [password, setPassword] = useState('');
  const [status, setStatus] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    setName(user?.name ?? '');
    setEmail(user?.email ?? '');
  }, [user]);

  if (!user || !token) {
    return <p>Loading profile...</p>;
  }

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setSubmitting(true);
    setStatus(null);
    setError(null);

    try {
      const payload: { name?: string; email?: string; password?: string } = {};
      if (name !== user.name) {
        payload.name = name;
      }
      if (email !== user.email) {
        payload.email = email;
      }
      if (password) {
        payload.password = password;
      }

      if (Object.keys(payload).length === 0) {
        setStatus('No changes to update');
        return;
      }

      await updateUser(token, user.id, payload);
      await refreshProfile();
      setPassword('');
      setStatus('Profile updated successfully');
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to update profile';
      setError(message);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="card">
      <h2>Your profile</h2>
      <form onSubmit={handleSubmit} className="form">
        <label htmlFor="name">Name</label>
        <input id="name" value={name} onChange={(event) => setName(event.target.value)} required />

        <label htmlFor="email">Email</label>
        <input id="email" type="email" value={email} onChange={(event) => setEmail(event.target.value)} required />

        <label htmlFor="password">New password</label>
        <input
          id="password"
          type="password"
          value={password}
          onChange={(event) => setPassword(event.target.value)}
          placeholder="Leave blank to keep current password"
          minLength={6}
        />

        {status && <p className="success">{status}</p>}
        {error && <p className="error">{error}</p>}

        <button type="submit" className="primary" disabled={submitting}>
          {submitting ? 'Saving...' : 'Save changes'}
        </button>
      </form>
    </div>
  );
};

export default ProfilePage;
