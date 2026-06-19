// =============================================================================
// TypeScript Interfaces for Logistics Microservices
// =============================================================================

export type TrackingStatus =
  | 'CREATED'
  | 'IN_HUB'
  | 'IN_TRANSIT'
  | 'OUT_DELIVERY'
  | 'DELIVERED'
  | 'FAILED'
  | 'RETURNED';

export type ServiceType = 'REGULAR' | 'EXPRESS' | 'SAMEDAY' | 'NEXTDAY';

export type UserRole = 'admin' | 'courier';

// Auth Models
export interface User {
  id: number;
  full_name: string;
  email: string;
  role: string;
  created_at: string;
  updated_at: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
}

export interface RegisterRequest {
  full_name: string;
  email: string;
  password: string;
}

export interface RegisterResponse {
  id: number;
  email: string;
  full_name: string;
  role: string;
  created_at: string;
}

export interface ValidateRequest {
  token: string;
}

export interface ValidateResponse {
  user_id: string;
  role: string;
}

// Pricing Models
export interface PricingRequest {
  origin: string;
  destination: string;
  weight: number;
  distance: number;
  service_type: ServiceType;
}

export interface PricingResult {
  total_cost: number;
  base_cost: number;
}

// Order Models
export interface OrderRequest {
  sender_name: string;
  recipient_name: string;
  origin: string;
  destination: string;
  weight: number;
  dimensions: string;
  item_type: string;
  service_type: ServiceType;
  distance: number;
}

export interface OrderResponse {
  order_id: string;
  resi_id: string;
  status: TrackingStatus;
  total_cost: number;
}

export interface ValidateResiResponse {
  status: 'valid' | string;
}

// Shipment Models
export interface Shipment {
  resi_id: string;
  status: TrackingStatus;
  hub_id: string;
  courier_id: string;
  proof_url: string;
  updated_at: string;
}

// Tracking Models
export interface TrackingRequest {
  resi_id: string;
  status: TrackingStatus;
  location: string;
  note: string;
}

export interface TrackingEvent {
  id: string;
  resi_id: string;
  status: TrackingStatus;
  location: string;
  note: string;
  created_at: string;
}

// Hub Models
export interface ScanRequest {
  resi_id: string;
  hub_id: string;
}

export interface SuccessResponse {
  status: string;
  message: string;
}

export interface ErrorResponse {
  error: string;
  message: string;
}

// Courier Models
export interface AssignRequest {
  resi_id: string;
  courier_id: string;
}

export interface DeliveryStatusRequest {
  resi_id: string;
  status: TrackingStatus;
  proof_url: string;
}

// Notification Models
export interface NotifyRequest {
  user_id: string;
  message: string;
  type: 'EMAIL' | 'SMS' | string;
}

export interface NotifyResponse {
  status: string;
  message: string;
  timestamp: string;
}
