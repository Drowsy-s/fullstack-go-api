import React, { useEffect, useState } from 'react';
import { fetchUsers } from '../api';
import { useAuth } from '../context/AuthContext';
import type { User } from '../types';

const DashboardPage: React.FC = () => {
  const { token, user } = useAuth();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;

    const loadUsers = async () => {
      if (!token) {
        setUsers([]);
        setLoading(false);
        return;
      }

      setLoading(true);
      setError(null);
      try {
        const { users: list } = await fetchUsers(token);
        if (active) {
          setUsers(list);
        }
      } catch (err) {
        if (active) {
          const message = err instanceof Error ? err.message : 'Failed to load users';
          setError(message);
        }
      } finally {
        if (active) {
          setLoading(false);
        }
      }
    };

    void loadUsers();

    return () => {
      active = false;
    };
  }, [token]);

  return (
    <div className="card">
      <div className="card-header">
        <div>
          <h2>User dashboard</h2>
          <p>Welcome back, {user?.name ?? 'friend'}.</p>
        </div>
      </div>

      {loading && <p>Loading users...</p>}
      {error && <p className="error">{error}</p>}

      {!loading && !error && (
        <div className="table-wrapper">
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>Name</th>
                <th>Email</th>
                <th>Created</th>
                <th>Updated</th>
              </tr>
            </thead>
            <tbody>
              {users.map((item) => (
                <tr key={item.id}>
                  <td>{item.id}</td>
                  <td>{item.name}</td>
                  <td>{item.email}</td>
                  <td>{new Date(item.createdAt).toLocaleString()}</td>
                  <td>{new Date(item.updatedAt).toLocaleString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};

export default DashboardPage;
