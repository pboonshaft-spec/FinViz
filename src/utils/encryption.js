import nacl from 'tweetnacl';
import { encodeBase64, decodeBase64, encodeUTF8, decodeUTF8 } from 'tweetnacl-util';

const KEY_STORAGE_PREFIX = 'finviz_e2e_';

/**
 * Generates a new key pair for the user
 * @returns {Object} { publicKey, secretKey, keyId }
 */
export function generateKeyPair() {
  const keyPair = nacl.box.keyPair();
  const keyId = generateKeyId();

  return {
    publicKey: encodeBase64(keyPair.publicKey),
    secretKey: encodeBase64(keyPair.secretKey),
    keyId,
  };
}

/**
 * Generates a random key ID
 * @returns {string}
 */
function generateKeyId() {
  const bytes = nacl.randomBytes(16);
  return encodeBase64(bytes).replace(/[+/=]/g, '').substring(0, 16);
}

/**
 * Stores keys securely in localStorage
 * @param {Object} keys - { publicKey, secretKey, keyId }
 */
export function storeKeys(keys) {
  localStorage.setItem(`${KEY_STORAGE_PREFIX}public`, keys.publicKey);
  localStorage.setItem(`${KEY_STORAGE_PREFIX}secret`, keys.secretKey);
  localStorage.setItem(`${KEY_STORAGE_PREFIX}keyId`, keys.keyId);
}

/**
 * Retrieves stored keys from localStorage
 * @returns {Object|null} { publicKey, secretKey, keyId } or null
 */
export function getStoredKeys() {
  const publicKey = localStorage.getItem(`${KEY_STORAGE_PREFIX}public`);
  const secretKey = localStorage.getItem(`${KEY_STORAGE_PREFIX}secret`);
  const keyId = localStorage.getItem(`${KEY_STORAGE_PREFIX}keyId`);

  if (!publicKey || !secretKey || !keyId) {
    return null;
  }

  return { publicKey, secretKey, keyId };
}

/**
 * Clears stored keys (used on logout)
 */
export function clearKeys() {
  localStorage.removeItem(`${KEY_STORAGE_PREFIX}public`);
  localStorage.removeItem(`${KEY_STORAGE_PREFIX}secret`);
  localStorage.removeItem(`${KEY_STORAGE_PREFIX}keyId`);
}

/**
 * Encrypts a message for a recipient
 * @param {string} message - Plain text message
 * @param {string} recipientPublicKey - Base64 encoded public key
 * @param {string} senderSecretKey - Base64 encoded secret key
 * @returns {Object} { encryptedContent, nonce }
 */
export function encryptMessage(message, recipientPublicKey, senderSecretKey) {
  const messageBytes = decodeUTF8(message);
  const nonce = nacl.randomBytes(nacl.box.nonceLength);

  const recipientPubKeyBytes = decodeBase64(recipientPublicKey);
  const senderSecKeyBytes = decodeBase64(senderSecretKey);

  const encrypted = nacl.box(
    messageBytes,
    nonce,
    recipientPubKeyBytes,
    senderSecKeyBytes
  );

  return {
    encryptedContent: encodeBase64(encrypted),
    nonce: encodeBase64(nonce),
  };
}

/**
 * Decrypts a message from a sender
 * @param {string} encryptedContent - Base64 encoded encrypted message
 * @param {string} nonce - Base64 encoded nonce
 * @param {string} senderPublicKey - Base64 encoded public key of sender
 * @param {string} recipientSecretKey - Base64 encoded secret key of recipient
 * @returns {string|null} Decrypted message or null if decryption fails
 */
export function decryptMessage(encryptedContent, nonce, senderPublicKey, recipientSecretKey) {
  try {
    const encryptedBytes = decodeBase64(encryptedContent);
    const nonceBytes = decodeBase64(nonce);
    const senderPubKeyBytes = decodeBase64(senderPublicKey);
    const recipientSecKeyBytes = decodeBase64(recipientSecretKey);

    const decrypted = nacl.box.open(
      encryptedBytes,
      nonceBytes,
      senderPubKeyBytes,
      recipientSecKeyBytes
    );

    if (!decrypted) {
      console.error('Failed to decrypt message');
      return null;
    }

    return encodeUTF8(decrypted);
  } catch (error) {
    console.error('Decryption error:', error);
    return null;
  }
}

/**
 * Initialize encryption for a user - generates keys if needed
 * @param {Function} registerPublicKey - API function to register public key
 * @returns {Object} { publicKey, secretKey, keyId }
 */
export async function initializeEncryption(registerPublicKey) {
  let keys = getStoredKeys();

  if (!keys) {
    // Generate new keys
    keys = generateKeyPair();
    storeKeys(keys);

    // Register public key with server
    try {
      await registerPublicKey({
        publicKey: keys.publicKey,
        keyId: keys.keyId,
      });
    } catch (error) {
      console.error('Failed to register public key:', error);
      // Don't throw - keys are still usable locally
    }
  }

  return keys;
}

/**
 * Cache for public keys of other users
 */
const publicKeyCache = new Map();

/**
 * Gets or fetches a user's public key
 * @param {number} userId - User ID
 * @param {Function} getPublicKey - API function to fetch public key
 * @returns {string|null} Base64 encoded public key
 */
export async function getUserPublicKey(userId, getPublicKey) {
  // Check cache first
  if (publicKeyCache.has(userId)) {
    return publicKeyCache.get(userId);
  }

  try {
    const keyData = await getPublicKey(userId);
    if (keyData && keyData.publicKey) {
      publicKeyCache.set(userId, keyData.publicKey);
      return keyData.publicKey;
    }
  } catch (error) {
    console.error('Failed to fetch public key for user', userId, error);
  }

  return null;
}

/**
 * Clears the public key cache
 */
export function clearPublicKeyCache() {
  publicKeyCache.clear();
}
