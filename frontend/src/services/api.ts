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
  notification: process.env.REACT_APP_NOTIFICATION_API_URL || 'http://localhost:8086',
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
// API Response Helpers
// =============================================================================

interface ApiEnvelope<TData> {
  success: boolean;
  data: TData;
  error?: {
    code?: string;
    message?: string;
  };
}

const unwrapResponse = <TData>(responseData: ApiEnvelope<TData> | TData): TData => {
  if (
    responseData &&
    typeof responseData === 'object' &&
    'success' in responseData &&
    'data' in responseData
  ) {
    const envelope = responseData as ApiEnvelope<TData>;
    if (!envelope.success) {
      throw new Error(envelope.error?.message || 'Permintaan API gagal.');
    }
    return envelope.data;
  }

  return responseData as TData;
};

const extractApiErrorMessage = (error: unknown): string => {
  if (axios.isAxiosError(error)) {
    const responseData = error.response?.data as Partial<ApiEnvelope<unknown>> | undefined;
    return (
      responseData?.error?.message ||
      error.response?.statusText ||
      error.message ||
      'Tidak dapat terhubung ke backend.'
    );
  }

  return error instanceof Error ? error.message : 'Terjadi kesalahan tidak diketahui.';
};

export const toApiErrorMessage = extractApiErrorMessage;

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
      client.defaults.headers.common.Authorization = `Bearer ${token}`;
    } else {
      delete client.defaults.headers.common.Authorization;
    }
  });
};

const persistedToken = localStorage.getItem('token');
if (persistedToken) {
  setAuthToken(persistedToken);
}

// =============================================================================
// API Service Methods
// =============================================================================

// --- Auth Service APIs ---

/**
 * Logs in a user using email and password.
 * @route POST /login (Auth Service)
 */
export const loginUser = async (data: T.LoginRequest): Promise<T.LoginResponse> => {
  try {
    const response = await authClient.post<ApiEnvelope<T.LoginResponse>>('/login', data);
    return unwrapResponse(response.data);
  } catch (error) {
    throw new Error(extractApiErrorMessage(error));
  }
};

/**
 * Registers a new user account.
 * @route POST /register (Auth Service)
 */
export const registerUser = async (data: T.RegisterRequest): Promise<T.RegisterResponse> => {
  try {
    const response = await authClient.post<ApiEnvelope<T.RegisterResponse>>('/register', data);
    return unwrapResponse(response.data);
  } catch (error) {
    throw new Error(extractApiErrorMessage(error));
  }
};

/**
 * Validates a JWT token and returns user details.
 * @route POST /auth/validate (Auth Service)
 */
export const validateToken = async (token: string): Promise<T.ValidateResponse> => {
  try {
    const response = await authClient.post<ApiEnvelope<T.ValidateResponse>>('/auth/validate', { token });
    return unwrapResponse(response.data);
  } catch (error) {
    throw new Error(extractApiErrorMessage(error));
  }
};

// --- Pricing Service APIs ---

/**
 * Calculates tariff pricing for a shipment.
 * @route POST /pricing (Pricing Service)
 */
export const calculatePricing = async (data: T.PricingRequest): Promise<T.PricingResult> => {
  try {
    const response = await pricingClient.post<ApiEnvelope<T.PricingResult>>('/pricing', data);
    return unwrapResponse(response.data);
  } catch (error) {
    throw new Error(extractApiErrorMessage(error));
  }
};

// --- Order Service APIs ---

/**
 * Creates a new logistics order.
 * @route POST /order (Order Service)
 */
export const createOrder = async (data: T.OrderRequest): Promise<T.OrderResponse> => {
  try {
    const response = await orderClient.post<ApiEnvelope<T.OrderResponse>>('/order', data);
    return unwrapResponse(response.data);
  } catch (error) {
    throw new Error(extractApiErrorMessage(error));
  }
};

/**
 * Checks if a specific resiID exists and is valid.
 * @route GET /orders/{resiID}/validate (Order Service)
 */
export const validateResi = async (resiID: string): Promise<T.ValidateResiResponse> => {
  try {
    const response = await orderClient.get<ApiEnvelope<T.ValidateResiResponse>>(
      `/orders/${encodeURIComponent(resiID)}/validate`
    );
    return unwrapResponse(response.data);
  } catch (error) {
    throw new Error(extractApiErrorMessage(error));
  }
};

// --- Hub Service APIs (Requires Admin Role) ---

/**
 * Scans a package in at a specific Hub.
 * @route POST /hub/scan-in (Hub Service)
 */
export const scanInHub = async (data: T.ScanRequest): Promise<T.SuccessResponse> => {
  try {
    const response = await hubClient.post<ApiEnvelope<T.SuccessResponse>>('/hub/scan-in', data);
    return unwrapResponse(response.data);
  } catch (error) {
    throw new Error(extractApiErrorMessage(error));
  }
};

/**
 * Scans a package out of a specific Hub.
 * @route POST /hub/scan-out (Hub Service)
 */
export const scanOutHub = async (data: T.ScanRequest): Promise<T.SuccessResponse> => {
  try {
    const response = await hubClient.post<ApiEnvelope<T.SuccessResponse>>('/hub/scan-out', data);
    return unwrapResponse(response.data);
  } catch (error) {
    throw new Error(extractApiErrorMessage(error));
  }
};

// --- Courier Service APIs ---

/**
 * Assigns a courier to deliver a package. (Requires Admin Role)
 * @route POST /courier/assign (Courier Service)
 */
export const assignCourier = async (data: T.AssignRequest): Promise<T.SuccessResponse> => {
  try {
    const response = await courierClient.post<ApiEnvelope<T.SuccessResponse>>('/courier/assign', data);
    return unwrapResponse(response.data);
  } catch (error) {
    throw new Error(extractApiErrorMessage(error));
  }
};

/**
 * Updates delivery status of a package with a photo proof. (Requires Courier Role)
 * @route POST /courier/delivery-status (Courier Service)
 */
export const updateDeliveryStatus = async (data: T.DeliveryStatusRequest): Promise<T.SuccessResponse> => {
  try {
    const response = await courierClient.post<ApiEnvelope<T.SuccessResponse>>('/courier/delivery-status', data);
    return unwrapResponse(response.data);
  } catch (error) {
    throw new Error(extractApiErrorMessage(error));
  }
};

// --- Tracking Service APIs ---

/**
 * Adds a tracking event manually or routes events.
 * @route POST /track (Tracking Service)
 */
export const addTrackingEvent = async (data: T.TrackingRequest): Promise<{ message: string }> => {
  try {
    const response = await trackingClient.post<ApiEnvelope<{ message: string }>>('/track', data);
    return unwrapResponse(response.data);
  } catch (error) {
    throw new Error(extractApiErrorMessage(error));
  }
};

/**
 * Adds a tracking event history log.
 * @route POST /events (Tracking Service)
 */
export const addTrackingEventLog = async (data: T.TrackingRequest): Promise<{ message: string }> => {
  try {
    const response = await trackingClient.post<ApiEnvelope<{ message: string }>>('/events', data);
    return unwrapResponse(response.data);
  } catch (error) {
    throw new Error(extractApiErrorMessage(error));
  }
};

// --- Notification Service APIs ---

/**
 * Sends notifications (email or SMS) to users.
 * @route POST /notify (Notification Service)
 */
export const sendNotification = async (data: T.NotifyRequest): Promise<T.NotifyResponse> => {
  try {
    const response = await notificationClient.post<ApiEnvelope<T.NotifyResponse>>('/notify', data);
    return unwrapResponse(response.data);
  } catch (error) {
    throw new Error(extractApiErrorMessage(error));
  }
};
