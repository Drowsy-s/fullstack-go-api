/* eslint-disable */
print('Configuring MongoDB user management collections...');

const databaseName = typeof _getEnv === 'function' ? _getEnv('MONGO_INITDB_DATABASE') || 'app' : 'app';
db = db.getSiblingDB(databaseName);

const usersCollection = db.getCollection('users');
usersCollection.createIndex({ email: 1 }, { unique: true, name: 'users_email_unique' });
usersCollection.createIndex({ isActive: 1 }, { name: 'users_isActive_idx' });
usersCollection.updateOne(
  { email: 'admin@example.com' },
  {
    $setOnInsert: {
      email: 'admin@example.com',
      passwordHash: '<replace-with-sha256-hash>',
      displayName: 'Administrator',
      roles: ['admin'],
      isActive: true,
      metadata: {},
      createdAt: new Date(),
      updatedAt: new Date()
    }
  },
  { upsert: true }
);
