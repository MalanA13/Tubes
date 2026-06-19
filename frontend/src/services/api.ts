import axios, { AxiosInstance } from 'axios';
import * as T from '../types';

// =============================================================================
// Microservice Base URL Configurations
// =============================================================================

export const SERVICES_BASE_URLS = {
  auth: process.env.REACT_APP_AUTH_API_URL || 'http://localhost:8080',
  order: process.env.REACT_APP_ORDER_API_URL || 'http://localhost:8081',
  pricing: process.env.REACT_APP_PRICING_API_URL || 'http://localhost:8082',
  tracking: process.env.REACT_APP_TRACKING_API_URL || 'http://localhost:8083',
  hub: process.env.REACT_APP_HUB_API_URL || 'http://localhost:8084',
  courier: process.env.REACT_APP_COURIER_API_URL || 'http://localhost:8085',
  notification: process.env.REACT_APP_NOTIFICATION_API_URL || 'http://localhost:8080',
};

// =============================================================================
// Axios Instance Initializations
// =============================================================================

const authClient: AxiosInstance = axios.create({
  baseURL: SERVICES_BASE_URLS.auth,
  headers: { 'Content-Type': 'application/json' },
});

const orderClient: AxiosInstance = axios.create({
  baseURL: SERVICES_BASE_URLS.order,
  headers: { 'Content-Type': 'application/json' },
});

const pricingClient: AxiosInstance = axios.create({
  baseURL: SERVICES_BASE_URLS.pricing,
  headers: { 'Content-Type': 'application/json' },
});

const trackingClient: AxiosInstance = axios.create({
  baseURL: SERVICES_BASE_URLS.tracking,
  headers: { 'Content-Type': 'application/json' },
});

const hubClient: AxiosInstance = axios.create({
  baseURL: SERVICES_BASE_URLS.hub,
  headers: { 'Content-Type': 'application/json' },
});

const courierClient: AxiosInstance = axios.create({
  baseURL: SERVICES_BASE_URLS.courier,
  headers: { 'Content-Type': 'application/json' },
});

const notificationClient: AxiosInstance = axios.create({
  baseURL: SERVICES_BASE_URLS.notification,
  headers: { 'Content-Type': 'application/json' },
});

// =============================================================================
// Authentication Header Helper
// =============================================================================

/**
 * Sets or removes the Bearer Authorization header for clients that require JWT authentication.
 * @param token JWT token string or null to clear authentication headers.
 */
export const setAuthToken = (token: string | null): void => {
  const protectedClients = [hubClient, courierClient];
  protectedClients.forEach((client) => {
    if (token) {
      client.defaults.headers.common['Authorization'] = `Bearer ${token}`;
    } else {
      delete client.defaults.headers.common['Authorization'];
    }
  });
};

// =============================================================================
// API Service Methods
// =============================================================================

// --- Auth Service APIs ---

/**
 * Logs in a user using email and password.
 * @route POST /login (Auth Service)
 */
export const loginUser = async (data: T.LoginRequest): Promise<T.LoginResponse> => {
  const response = await authClient.post<T.LoginResponse>('/login', data);
  return response.data;
};

/**
 * Registers a new user account.
 * @route POST /register (Auth Service)
 */
export const registerUser = async (data: T.RegisterRequest): Promise<T.RegisterResponse> => {
  const response = await authClient.post<T.RegisterResponse>('/register', data);
  return response.data;
};

/**
 * Validates a JWT token and returns user details.
 * @route POST /auth/validate (Auth Service)
 */
export const validateToken = async (token: string): Promise<T.validateResponse> => {
  const response = await authClient.post<T.validateResponse>('/auth/validate', { token });
  return response.data;
};

// --- Pricing Service APIs ---

/**
 * Calculates tariff pricing for a shipment.
 * @route POST /pricing (Pricing Service)
 */
export const calculatePricing = async (data: T.PricingRequest): Promise<T.PricingResult> => {
  const response = await pricingClient.post<T.PricingResult>('/pricing', data);
  return response.data;
};

// --- Order Service APIs ---

/**
 * Creates a new logistics order.
 * @route POST /order (Order Service)
 */
export const createOrder = async (data: T.OrderRequest): Promise<T.OrderResponse> => {
  const response = await orderClient.post<T.OrderResponse>('/order', data);
  return response.data;
};

/**
 * Checks if a specific resiID exists and is valid.
 * @route GET /orders/{resiID}/validate (Order Service)
 */
export const validateResi = async (resiID: string): Promise<T.ValidateResiResponse> => {
  const response = await orderClient.get<T.ValidateResiResponse>(`/orders/${encodeURIComponent(resiID)}/validate`);
  return response.data;
};

// --- Hub Service APIs (Requires Admin Role) ---

/**
 * Scans a package in at a specific Hub.
 * @route POST /hub/scan-in (Hub Service)
 */
export const scanInHub = async (data: T.ScanRequest): Promise<T.SuccessResponse> => {
  const response = await hubClient.post<T.SuccessResponse>('/hub/scan-in', data);
  return response.data;
};

/**
 * Scans a package out of a specific Hub.
 * @route POST /hub/scan-out (Hub Service)
 */
export const scanOutHub = async (data: T.ScanRequest): Promise<T.SuccessResponse> => {
  const response = await hubClient.post<T.SuccessResponse>('/hub/scan-out', data);
  return response.data;
};

// --- Courier Service APIs ---

/**
 * Assigns a courier to deliver a package. (Requires Admin Role)
 * @route POST /courier/assign (Courier Service)
 */
export const assignCourier = async (data: T.AssignRequest): Promise<T.SuccessResponse> => {
  const response = await courierClient.post<T.SuccessResponse>('/courier/assign', data);
  return response.data;
};

/**
 * Updates delivery status of a package with a photo proof. (Requires Courier Role)
 * @route POST /courier/delivery-status (Courier Service)
 */
export const updateDeliveryStatus = async (data: T.DeliveryStatusRequest): Promise<T.SuccessResponse> => {
  const response = await courierClient.post<T.SuccessResponse>('/courier/delivery-status', data);
  return response.data;
};

// --- Tracking Service APIs ---

/**
 * Adds a tracking event manually or routes events.
 * @route POST /track (Tracking Service)
 */
export const addTrackingEvent = async (data: T.TrackingRequest): Promise<{ message: string }> => {
  const response = await trackingClient.post<{ message: string }>('/track', data);
  return response.data;
};

/**
 * Adds a tracking event history log.
 * @route POST /events (Tracking Service)
 */
export const addTrackingEventLog = async (data: T.TrackingRequest): Promise<{ message: string }> => {
  const response = await trackingClient.post<{ message: string }>('/events', data);
  return response.data;
};

// --- Notification Service APIs ---

/**
 * Sends notifications (email or SMS) to users.
 * @route POST /notify (Notification Service)
 */
export const sendNotification = async (data: T.NotifyRequest): Promise<T.NotifyResponse> => {
  const response = await notificationClient.post<T.NotifyResponse>('/notify', data);
  return response.data;
};
