import React, { useState, useEffect } from 'react';
import { useSnackbar } from 'notistack';
import { Key, Plus, Trash2, X } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import Card from '@monetr/interface/components/Card';
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@monetr/interface/components/Dialog';
import { Input } from '@monetr/interface/components/Input';
import MSelect, { Value } from '@monetr/interface/components/MSelect';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@monetr/interface/components/Table';
import request from '@monetr/interface/util/request';

interface APIKey {
  apiKeyId: number | string;
  name: string;
  createdAt: string;
  lastUsedAt?: string;
  expiresAt?: string;
  isActive: boolean;
  userId: string;
}

export default function APIKeysCard(): JSX.Element {
  const [keys, setKeys] = useState<APIKey[]>([]);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [newKey, setNewKey] = useState<string | null>(null);
  const [keyName, setKeyName] = useState('');
  const [expiration, setExpiration] = useState<Value<string> | null>(null);
  const { enqueueSnackbar } = useSnackbar();

  const loadKeys = async () => {
    try {
      const result = await request().get('/security/api-keys');
      console.log('API Keys response:', result.data);
      // Ensure we have a valid array of keys
      if (Array.isArray(result.data)) {
        setKeys(result.data);
      } else {
        console.error('Unexpected API response format:', result.data);
        setKeys([]);
      }
    } catch (error) {
      console.error('API Keys error:', error);
      enqueueSnackbar('Failed to load API keys', { variant: 'error' });
    }
  };

  useEffect(() => {
    loadKeys();
  }, []);

  const handleRevoke = async (apiKeyId: number | string) => {
    try {
      const result = await request().delete(`/security/api-keys/${apiKeyId}`);
      console.log('Revoke API Key response:', result.data);
      await loadKeys();
      enqueueSnackbar('API key revoked', { variant: 'success' });
    } catch (error) {
      console.error('Revoke API Key error:', error);
      enqueueSnackbar('Failed to revoke API key', { variant: 'error' });
    }
  };

  const handleCreateKey = async () => {
    try {
      // Calculate expiration date based on selected option
      let expiresAt: Date | null = null;
      
      if (expiration && expiration.value !== 'none') {
        expiresAt = new Date();
        const days = {
          '1day': 1,
          '1week': 7,
          '1month': 30,
          '1year': 365,
        }[expiration.value] || 0;
        
        expiresAt.setDate(expiresAt.getDate() + days);
      }

      const payload = { 
        name: keyName,
        ...(expiresAt && { expiresAt: expiresAt.toISOString() })
      };

      const result = await request().post('/security/api-keys', payload);
      console.log('Create API Key response:', result.data);
      if (result.data && result.data.key) {
        setNewKey(result.data.key);
        setIsDialogOpen(false);
        setKeyName('');
        setExpiration(null);
        loadKeys();
      } else {
        console.error('Unexpected API response format:', result.data);
        enqueueSnackbar('Received invalid response when creating API key', { variant: 'error' });
      }
    } catch (error) {
      console.error('Create API Key error:', error);
      enqueueSnackbar('Failed to create API key', { variant: 'error' });
    }
  };

  return (
    <Card>
      <div className="flex justify-between items-center mb-4">
        <div className="flex items-center gap-2">
          <div className="border-dark-monetr-border rounded border w-fit p-2 bg-dark-monetr-background-subtle">
            <Key size={20} />
          </div>
          <h2 className="text-xl font-semibold text-dark-monetr-content-emphasis">API Keys</h2>
        </div>
        <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
          <DialogTrigger asChild>
            <Button variant="primary">
              <Plus size={16} />
              Create Key
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create API Key</DialogTitle>
            </DialogHeader>
            <div className="py-4 space-y-4">
              <Input
                autoFocus
                placeholder="Key Name"
                className="w-full"
                value={keyName}
                onChange={e => setKeyName(e.target.value)}
              />
              
              <div className="space-y-2">
                <label className="text-sm font-medium text-dark-monetr-content-emphasis">
                  Expiration
                </label>
                <MSelect
                  placeholder="Select expiration period"
                  options={[
                    { label: 'No Expiration', value: 'none' },
                    { label: '1 Day', value: '1day' },
                    { label: '1 Week', value: '1week' },
                    { label: '1 Month (30 days)', value: '1month' },
                    { label: '1 Year (365 days)', value: '1year' },
                  ]}
                  value={expiration}
                  onChange={(option) => setExpiration(option as Value<string>)}
                />
              </div>
            </div>
            <DialogFooter>
              <DialogClose asChild>
                <Button variant="secondary">Cancel</Button>
              </DialogClose>
              <Button 
                variant="primary" 
                onClick={handleCreateKey} 
                disabled={!keyName}
              >
                Create
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {newKey && (
        <div className="mb-4 p-4 bg-dark-monetr-background-subtle border border-dark-monetr-border rounded-lg text-dark-monetr-content-emphasis">
          <p className="font-semibold">New API Key Created</p>
          <p className="break-all font-mono bg-dark-monetr-background p-2 rounded mt-2 border border-dark-monetr-border">{newKey}</p>
          <p className="text-sm text-dark-monetr-content mt-2 font-medium">
            Make sure to copy this key now. You won't be able to see it again!
          </p>
          <Button
            variant="secondary"
            className="mt-2"
            onClick={() => setNewKey(null)}
          >
            Dismiss
          </Button>
        </div>
      )}

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Created</TableHead>
              <TableHead>Last Used</TableHead>
              <TableHead>Expires</TableHead>
              <TableHead className="w-[80px]">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {keys.map(key => (
              <TableRow key={key.apiKeyId}>
                <TableCell>{key.name}</TableCell>
                <TableCell>{new Date(key.createdAt).toLocaleDateString()}</TableCell>
                <TableCell>{key.lastUsedAt ? new Date(key.lastUsedAt).toLocaleDateString() : 'Never'}</TableCell>
                <TableCell>
                  {key.expiresAt 
                    ? new Date(key.expiresAt).toLocaleDateString(undefined, {
                        year: 'numeric',
                        month: 'short',
                        day: 'numeric'
                      })
                    : 'Never'}
                </TableCell>
                <TableCell>
                  <Button
                    variant="text"
                    onClick={() => handleRevoke(typeof key.apiKeyId === 'string' ? parseInt(key.apiKeyId) : key.apiKeyId)}
                    disabled={!key.isActive}
                    className="p-1"
                  >
                    <Trash2 size={16} className="text-red-500" />
                  </Button>
                </TableCell>
              </TableRow>
            ))}
            {keys.length === 0 && (
              <TableRow>
                <TableCell colSpan={5} className="h-24 text-center text-dark-monetr-content">
                  No API keys found
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
    </Card>
  );
}
