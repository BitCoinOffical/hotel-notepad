export interface Guest {
  id: string;
  user_id: string;
  first_name: string;
  mid_name: string;
  last_name: string;
  note: string | null;
  phone: number;
  is_delete: boolean;
  arrival_date: string;
  departure_date: string;
  created_at: string;
  updated_at: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest extends LoginRequest {
  password_retry: string;
}

export interface CreateGuestRequest {
  first_name: string;
  mid_name: string;
  last_name: string;
  note: string | null;
  phone: number;
  arrival_date: string;
  departure_date: string;
}

export interface UpdateGuestRequest {
  first_name: string;
  mid_name: string;
  last_name: string;
  note: string;
  phone: number;
  arrival_date: string;
  departure_date: string;
}

export interface GuestFormValues {
  firstName: string;
  midName: string;
  lastName: string;
  note: string;
  phone: string;
  arrivalDate: string;
  departureDate: string;
}
